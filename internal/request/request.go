package request

import (
	"io"
	"strings"
	"fmt"
	"errors"
	"strconv"
	"github.com/aringq10/http-go-server/internal/headers"
)

// Request states
type State int

const (
	initialized State = iota
	parsingRequestLine
	parsingHeaders
	parsingBody
	done
)

const crlf = "\r\n"
const bufferSize = 4096

var httpMethods = map[string]struct{}{
    "GET":     {},
    "HEAD":    {},
    "POST":    {},
    "PUT":     {},
    "DELETE":  {},
    "CONNECT": {},
    "OPTIONS": {},
    "TRACE":   {},
    "PATCH":   {},
}

type Request struct {
	RequestLine RequestLine
	Headers headers.Headers
	Body []byte
	state State
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func NewRequest() Request {
	req := Request{}
	req.Headers = make(headers.Headers)
	return req
}

func validMethod(method string) bool {
    _, ok := httpMethods[method]
    return ok
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := NewRequest()
	req.state = initialized
	readToIndex := 0 // Index up to which the buffer is filled
	buf := make([]byte, bufferSize)

	for req.state != done {
		if readToIndex >= len(buf) {
			extBuf := make([]byte, len(buf) * 2)
			copy(extBuf, buf)
			buf = extBuf
		}

		n, err := reader.Read(buf[readToIndex:])

		if err != nil {
			if errors.Is(err, io.EOF) {
				for req.state != done{
					if readToIndex == 0 {
						switch req.state {
						case initialized:
							return nil, errors.New("error: missing request line\n")
						case parsingHeaders:
							return nil, errors.New("error: missing end of headers\n")
						case parsingBody:
							return nil, errors.New("error: body smaller than reported content length\n")
						}
					}
					n, err = req.parse(buf[:readToIndex])
					if err != nil {
						return nil, err
					}

					tmpBuf := buf[n:]
					copy(buf, tmpBuf)
					readToIndex -= n
				}

				break
			}
			return  nil, fmt.Errorf("error: %s\n", err.Error())
		}
		readToIndex += n

		n, err = req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}
		tmpBuf := buf[n:]
		copy(buf, tmpBuf)
		readToIndex -= n
	}

	return &req, nil
}

func (r *Request) parse(data []byte) (totalBytesParsed int, err error) {
	for r.state != done {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			break
		}
		totalBytesParsed += n
	}

	return
}


func (r *Request) parseSingle(data []byte) (n int, err error) {
	switch r.state {
	case initialized, parsingRequestLine:
		r.state = parsingRequestLine
		r.RequestLine, n, err = parseRequestLine(data)
		if err != nil {
			return 0, fmt.Errorf("error while parsing request line:\n    %v", err.Error())
		}
		if n != 0 {
			r.state = parsingHeaders
		}
		return
	case parsingHeaders:
		totalBytesParsed := 0
		finished := false

		for r.state != parsingBody {
			n, finished, err = r.Headers.Parse(data[totalBytesParsed:])
			if err != nil {
				return 0, fmt.Errorf("error while parsing headers:\n    %v", err.Error())
			}
			if n == 0 {
				break
			}
			if finished {
				r.state = parsingBody
			}

			totalBytesParsed += n
		}

		return totalBytesParsed, nil
	case parsingBody:
		lengthStr := r.Headers.Get("Content-Length")

		length, errConv := strconv.Atoi(lengthStr)
		if (errConv != nil || length == 0) {
			if len(data) != 0 {
				fmt.Println("No Content-Length but Body Exists")
			}
			r.state = done
			return 0, nil
		}

		if len(r.Body) < length && len(data) == 0 {
			return 0, nil
		}

		r.Body = append(r.Body, data...)
		if len(r.Body) > length {
			r.state = done
			return 0, errors.New("error: body larger than reported content length\n")
		} else if len(r.Body) == length {
			r.state = done
		}
		return len(data), nil
	case done:
		return 0, errors.New("error: trying to read data in a done state\n")
	default:
		return 0, errors.New("error: unknown parser state\n")
	}
}

func parseRequestLine(data []byte) (reqLine RequestLine, n int, err error) {
	tmpSplice := strings.Split(string(data), crlf)

	if len(tmpSplice) == 1 {
		return RequestLine{}, 0, nil
	}
	reqLineString := tmpSplice[0]

	n = len(reqLineString) + 2
	err = nil

	fields := strings.Fields(reqLineString)
	if len(fields) != 3 {
		return RequestLine{}, 0, errors.New("Request line doesn't have 3 fields")
	}
	method := fields[0]
	requestTarget := fields[1]
	httpVersion := fields[2]

	if method != strings.ToUpper(method) {
		return RequestLine{}, 0, errors.New("Invalid request method")
	}
	if !validMethod(method) {
		return RequestLine{}, 0, errors.New("Invalid request method")
	}

	reqLine.Method = method

	reqLine.RequestTarget = requestTarget

	tmpSplice = strings.Split(fields[2], "/")
	if len(tmpSplice) != 2 {
		return RequestLine{}, 0, errors.New("Invalid HTTP version")
	}
	protocol := tmpSplice[0]
	httpVersion = tmpSplice[1]
	if protocol != "HTTP" {
		return RequestLine{}, 0, errors.New("Invalid HTTP version")
	}
	tmpSplice = strings.Split(httpVersion, ".")
	if len(tmpSplice) != 2 {
		return RequestLine{}, 0, errors.New("Invalid HTTP version")
	}
	majorVersion := tmpSplice[0]
	minorVersion := tmpSplice[1]
	if majorVersion != "1" || minorVersion != "1" {
		return RequestLine{}, 0, errors.New("Invalid HTTP version")
	}
	reqLine.HttpVersion = httpVersion

	return
}
