package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/RegistersNinja/httpfromtcp/internal/headers"
	"github.com/RegistersNinja/httpfromtcp/internal/request"
	"github.com/RegistersNinja/httpfromtcp/internal/response"
	"github.com/RegistersNinja/httpfromtcp/internal/server"
)

const port = 42069

func respond400() []byte {
	return []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
}

func respond500() []byte {
	return []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
}

func respond200() []byte {
	return []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
}

func main() {
	var (
		handler server.Handler
		srv     *server.Server
		err     error
		sigChan chan os.Signal
	)

	handler = func(w *response.Writer, req *request.Request) {
		var (
			status  response.StatusCode
			headers headers.Headers
			body    []byte
		)

		status = response.StatusOK
		body = respond200()
		headers = response.GetDefaultHeaders(len(body))

		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			status = response.StatusBadRequest
			body = respond400()
		case "/myproblem":
			status = response.StatusInternalServerError
			body = respond500()
		}

		headers.OverrideHeaderValue("Content-Length", strconv.Itoa(len(body)))
		w.WriteStatusLine(status)
		w.WriteHeaders(headers)
		w.WriteBody(body)
	}

	srv, err = server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()
	log.Println("Server started on port", port)

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
