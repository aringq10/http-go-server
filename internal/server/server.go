package server

import (
	"fmt"
	"net"

	"github.com/aringq10/http-go-server/internal/request"
	"github.com/aringq10/http-go-server/internal/response"
)

type Server struct {
    handler Handler
    listener net.Listener
    closed bool
}

type Handler func(w *response.Writer, req *request.Request)

func Serve(port uint16, handler Handler) (*Server, error) {
    listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        return nil, err
    }

    s := &Server{
        handler: handler,
        listener: listener,
        closed: false,
    }

    go s.listen()

    return s, nil
}

func (s *Server) Close() error {
    s.closed = true
    return s.listener.Close()
}

func (s *Server) listen() {
    for {
        conn, err := s.listener.Accept()

        if s.closed {
            return
        }

        if err != nil {
            fmt.Printf("error establishing connection: %s\n", err.Error())
            continue
        }

        go s.handle(conn)
    }
}

func (s *Server) handle(conn net.Conn) {
    defer conn.Close()

    responseWriter := response.Writer{ Writer: conn }
    req, err := request.RequestFromReader(conn)

    if err != nil {
        body := fmt.Sprintf("That's such a bad request tho...\n  %v\n", err.Error())
        responseWriter.WriteStatusLine(400)
        h := response.GetDefaultHeaders(0)
        h.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
        responseWriter.WriteHeaders(h)
        responseWriter.WriteBody([]byte(body))
        return
    }

    s.handler(&responseWriter, req)
}
