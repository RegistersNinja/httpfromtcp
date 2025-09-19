package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/RegistersNinja/httpfromtcp/internal/request"
	"github.com/RegistersNinja/httpfromtcp/internal/response"
	"github.com/RegistersNinja/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	var (
		handler server.Handler
		srv *server.Server
		err error
		sigChan chan os.Signal
	)

	handler = func(w io.Writer, req *request.Request) *server.HandlerError {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			return &server.HandlerError{StatusCode: response.StatusBadRequest, Message: "Your problem is not my problem\n"}
		case "/myproblem":
			return &server.HandlerError{StatusCode: response.StatusInternalServerError, Message: "Woopsie, my bad\n"}
		default:
			_, _ = io.WriteString(w, "All good, frfr\n")
			return nil
		}
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