
package main

import (
	"fmt"
	"log"
	"net"
	"github.com/aringq10/http-go-server/internal/request"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("error listening for TCP traffic: %s\n", err.Error())
	}
	defer listener.Close()
	
	fmt.Println("Listening for TCP traffic on", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("error: %s\n", err.Error())
		}
		fmt.Println("Accepted connection from", conn.RemoteAddr())

		req, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("Request Line:")
			fmt.Println(" - Method:", req.RequestLine.Method)
			fmt.Println(" - Target:", req.RequestLine.RequestTarget)
			fmt.Println(" - Version:", req.RequestLine.HttpVersion)
		}

		fmt.Println("Connection to ", conn.RemoteAddr(), "closed")
	}
}

