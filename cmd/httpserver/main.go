package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

const response404 = `<html>
  <head>
    <title>404 Not Found</title>
  </head>
  <body>
    <h1>Not Found</h1>
    <p>You should try looking elsewhere.</p>
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

func reqHandler(w *response.Writer, req *request.Request) {
    h := response.GetDefaultHeaders(0)
    target := req.RequestLine.RequestTarget
    h.Replace("Content-Type", "text/html")

    var status int
    var body []byte

    if target == "/yourproblem" {
        status = 400
        body = []byte(response400)
    } else if target == "/myproblem" {
        status = 500
        body = []byte(response500)
    } else if strings.HasPrefix(target, "/httpbin/stream") {
        requestURL := "https://httpbin.org" + strings.TrimPrefix(target, "/httpbin")

        resp, err := http.Get(requestURL)
        if err != nil {
            status = 500
            body = []byte(response500)
            fmt.Println(err.Error())
        } else {
            w.WriteChunksFromReader(resp.Body, h)
            resp.Body.Close()
            return
        }
    } else if strings.HasPrefix(target, "/video") {
        f, err := os.Open("assets" + target)
        info, infoErr := os.Stat("assets" + target)
        if err != nil || infoErr != nil || info.IsDir() {
            status = 404
            body = []byte(response404)
            var errMsg string
            if err != nil {
                errMsg = err.Error()
            } else if infoErr != nil {
                errMsg = infoErr.Error()
            } else {
                errMsg = fmt.Sprintf("trying to open directory %v", "assets" + target)
            }
            fmt.Println(errMsg)
            f.Close()
        } else {
            w.WriteChunksFromReader(f, h)
            f.Close()
            return
        }

    } else {
        status = 200
        body = []byte(response200)
    }

    w.WriteHttpMessage(status, h, body)
}

func main() {
    server, err := server.Serve(port, reqHandler)

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
