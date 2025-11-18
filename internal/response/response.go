package response

import (
	"crypto/sha256"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/aringq10/http-go-server/internal/headers"
)


var reasonPhrases = map[int]string{
    200: "OK",
    400: "Bad Request",
    404: "Not Found",
    500: "Internal Server Error",
}

type Writer struct {
    Writer io.Writer
}

func NewWriter(writer io.Writer) *Writer {
    return &Writer{ Writer: writer }
}

func (w *Writer) WriteStatusLine(statusCode int) error {
    reasonPhrase, ok := reasonPhrases[statusCode]
    if !ok {
        return fmt.Errorf("unrecognized status code %v", statusCode)
    }

    statusLine := fmt.Sprintf("HTTP/1.1 %v %v\r\n", statusCode, reasonPhrase)

    _, err := w.Writer.Write([]byte(statusLine))

    return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
    b := []byte{}

    for key, value := range headers {
        b = fmt.Appendf(b, "%v: %v\r\n", key, value)
    }
    b = fmt.Append(b, "\r\n")

    _, err := w.Writer.Write(b)

    return err
}

func (w *Writer) WriteBody(body []byte) (int, error) {
    return w.Writer.Write(body)
}

func (w *Writer) WriteHttpMessage(statusCode int, h headers.Headers, body []byte) error {
    err := w.WriteStatusLine(statusCode)
    if err != nil {
        return err
    }
    h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
    err = w.WriteHeaders(h)
    if err != nil {
        return err
    }
    _, err = w.WriteBody(body)

    return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
    h := headers.NewHeaders()
    h.Set("Content-Length", strconv.Itoa(contentLen))
    h.Set("Connection", "close")
    h.Set("Content-Type", "text/plain")

    return h
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
    numLine := fmt.Sprintf("%X\r\n", len(p))
    p = fmt.Append(p, "\r\n")
    chunk := append([]byte(numLine), p...)

    return w.WriteBody(chunk)
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
    return w.WriteBody([]byte("0\r\n"))
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
    trailerString := h.Get("Trailer")
    trailerHeaders := headers.NewHeaders()

    trailers := strings.Split(trailerString, ", ")

    for _, trailer := range trailers {
        value := h.Get(trailer)
        if value == "" {
            continue
        }
        trailerHeaders[trailer] = value
    }

    return w.WriteHeaders(trailerHeaders)
}

func (w *Writer) WriteChunksFromReader(reader io.Reader, h headers.Headers) {
    h.Remove("Content-Length")
    h.Set("Transfer-Encoding", "chunked")
    h.Replace("Content-Type", "video/mp4")
    h.Set("Trailer", "X-Content-SHA256, X-Content-Length")
    w.WriteStatusLine(200)
    w.WriteHeaders(h)
    bytesRead := 0
    wholeResp := []byte{}

    for {
        buf := make([]byte, 32)
        n, err := reader.Read(buf)
        if err != nil {
            break
        }
        bytesRead += n
        wholeResp = append(wholeResp, buf[:n]...)

        w.WriteChunkedBody(buf[:n])
    }

    w.WriteChunkedBodyDone()

    h.Set("X-Content-SHA256", fmt.Sprintf("%X", sha256.Sum256(wholeResp)))
    h.Set("X-Content-Length", fmt.Sprintf("%d", bytesRead))

    w.WriteTrailers(h)


}
