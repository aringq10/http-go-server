package main

import (
	"log"
    "fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/aringq10/http-go-server/internal/request"
	"github.com/aringq10/http-go-server/internal/response"
	"github.com/aringq10/http-go-server/internal/server"
)

const port = 42069

const response400 = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

const response500 = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

const response200 = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

func main() {
    server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
        h := response.GetDefaultHeaders(0)
        h.Replace("Content-Type", "text/html")
        var status int
        var body []byte

        if req.RequestLine.RequestTarget == "/yourproblem" {
            h.Replace("Content-Length", fmt.Sprintf("%d", len(response400)))
            status = 400
            body = []byte(response400)
        } else if req.RequestLine.RequestTarget == "/myproblem" {
            h.Replace("Content-Length", fmt.Sprintf("%d", len(response500)))
            status = 500
            body = []byte(response500)
        } else {
            h.Replace("Content-Length", fmt.Sprintf("%d", len(response200)))
            status = 200
            body = []byte(response200)
        }



        err := w.WriteStatusLine(status)
        if err != nil {
            fmt.Println(err.Error())
            return
        }
        err = w.WriteHeaders(h)
        if err != nil {
            fmt.Println(err.Error())
            return
        }
        _, err = w.WriteBody(body)
        if err != nil {
            fmt.Println(err.Error())
            return
        }
    })

    if err != nil {
        log.Fatalf("Error starting server: %v\n", err)
    }
    defer server.Close()
    log.Println("Server started on port", port)

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan
    log.Println("Server gracefully stopped")
}
