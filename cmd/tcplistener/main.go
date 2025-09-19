package main

import (
	"fmt"
	"net"
	"os"

	"github.com/RegistersNinja/httpfromtcp/internal/request"
)

const (
	ExitError   int    = 1
	localAddr   string = "127.0.0.1:42069"
)

func main() {
	var (
		listener net.Listener
		conn     net.Conn
		r        *request.Request
		err      error
	)

	listener, err = net.Listen("tcp", localAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to listen: %v\n", err)
		os.Exit(ExitError)
	}

	defer listener.Close()

	for {
		conn, err = listener.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to connect: %v\n", err)
			os.Exit(ExitError)
		}

		r, err = request.RequestFromReader(conn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse the request: %v\n", err)
			os.Exit(ExitError)
		}

		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)

	}

}
