package request

import (
	"io"
	"strings"
	"fmt"
	"errors"
)

// Request states
const initialized = 1
const done = 2

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

func validMethod(method string) bool {
    _, ok := httpMethods[method]
    return ok
}

type Request struct {
	RequestLine RequestLine
	FieldLines []FieldLine
	state int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type FieldLine struct {
	key string
	value string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	var req Request
	req.state = initialized
	readToIndex := 0 // Index up to which the buffer is filled
	bytesRead := 0
	buf := make([]byte, bufferSize)

	for req.state == initialized {
		if readToIndex >= len(buf) {
			extBuf := make([]byte, len(buf) * 2)
			copy(extBuf, buf)
			buf = extBuf
		}
		n, err := reader.Read(buf[readToIndex:])

		if err != nil {
			if errors.Is(err, io.EOF) {
				req.state = done
				if bytesRead == 0 {
					return nil, errors.New("error: couldn't find a CRLF sequence\n")
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
		bytesRead += n
	}

	return &req, nil
}

func (r *Request) parse(data []byte) (n int, err error) {
	if r.state == done {
		return 0, errors.New("error: trying to read data in a done state\n")
	}
	if r.state != initialized {
		return 0, errors.New("error: unknown state\n")
	}

	r.RequestLine, n, err = parseRequestLine(data)
	if err != nil {
		return 0, fmt.Errorf("error while parsing request line: %v", err.Error())
	}
	if n != 0 {
		r.state = done
	}
	return
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
