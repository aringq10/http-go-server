package server

import (
	"fmt"
	"net"
	"log"
	"strconv"
	"github.com/aringq10/http-go-server/internal/request"
)

const response = "HTTP/1.1 200 OK\r\n" +
				 "Content-Type: text/plain\r\n" +
				 "Content-Length: 13\r\n" +
				 "\r\n" +
				 "Hello World!\n"

const (
	serverDown int = iota
	serverRunning
)

type Server struct {
	listener net.Listener
	state int
}

func Serve(port int) (*Server, error) {
	s := Server{}
	var err error = nil

	s.listener, err = net.Listen("tcp", "localhost:" + strconv.Itoa(port))
	if err != nil {
		log.Fatalf("error listening for TCP traffic: %s\n", err.Error())
	}
	s.state = serverRunning
	go s.listen()

	return &s, err
}

func (s *Server) Close() error {
	s.state = serverDown
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		if s.state == serverDown {
			break
		}
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Printf("error establishing connection: %s\n", err.Error())
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	fmt.Println("Accepted connection from", conn.RemoteAddr())

	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Request Line:")
		fmt.Println(" - Method:", req.RequestLine.Method)
		fmt.Println(" - Target:", req.RequestLine.RequestTarget)
		fmt.Println(" - Version:", req.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for key, value := range req.Headers {
			fmt.Printf(" - %v: %v\n", key, value)
		}

		fmt.Println("Body:")
		fmt.Println(string(req.Body))
	}

	conn.Write([]byte(response))

	conn.Close()
	fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
}
