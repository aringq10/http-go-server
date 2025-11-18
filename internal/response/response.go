package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/aringq10/http-go-server/internal/headers"
)

const (
    WriterStateStatusLine int = iota
    WriterStateHeaders
    WriterStateBody
    WriterStateDone
)

var reasonPhrases = map[int]string{
    200: "OK",
    400: "Bad Request",
    500: "Internal Server Error",
}

type Writer struct {
    Writer io.Writer
    State int
}

func (w *Writer) WriteStatusLine(statusCode int) error {
    if w.State != WriterStateStatusLine {
        return fmt.Errorf("trying to write response status line while in %v state", w.State)
    }
    reasonPhrase, ok := reasonPhrases[statusCode]
    if !ok {
        return fmt.Errorf("unrecognized status code %v", statusCode)
    }

    statusLine := fmt.Sprintf("HTTP/1.1 %v %v\r\n", statusCode, reasonPhrase)

    _, err := w.Writer.Write([]byte(statusLine))
    if err == nil {
        w.State = WriterStateHeaders
    }
    return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
    if w.State != WriterStateHeaders {
        return fmt.Errorf("trying to write response headers while in %v state", w.State)
    }
    b := []byte{}

    for key, value := range headers {
        b = fmt.Appendf(b, "%v: %v\r\n", key, value)
    }
    b = fmt.Append(b, "\r\n")

    _, err := w.Writer.Write(b)
    if err == nil {
        w.State = WriterStateBody
    }

    return err
}
func (w *Writer) WriteBody(p []byte) (int, error) {
    if w.State != WriterStateBody {
        return 0, fmt.Errorf("trying to write response body while in %v state", w.State)
    }

    n, err := w.Writer.Write(p)
    if err == nil {
        w.State = WriterStateDone
    }

    return n, err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
    h := headers.NewHeaders()
    h.Set("Content-Length", strconv.Itoa(contentLen))
    h.Set("Connection", "close")
    h.Set("Content-Type", "text/plain")

    return h
}

