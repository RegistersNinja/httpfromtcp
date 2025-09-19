package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

const (
	bytesToRead int = 8
	emptyRead   int = 0
	ExitError   int = 1
	channelSize int = 1
	localAddr string = "127.0.0.1:42069"
)

func getLinesChannel(conn net.Conn) <-chan string {
	out := make(chan string, channelSize)

	go func() {
		defer conn.Close()
		defer close(out)

		var (
			buffer       []byte
			line         string
			newLineSplit []string
		)

		buffer = make([]byte, bytesToRead)
		for {
			bytesRead, err := conn.Read(buffer)
			if err != nil {
				if err != io.EOF {
					fmt.Fprintf(os.Stderr, "failed to read from file: %v\n", err)
				}
				break
			}
			if bytesRead > emptyRead {
				line += string(buffer[:bytesRead])
			}

			newLineSplit = strings.Split(line, "\n")
			for i := 0; i < len(newLineSplit)-1; i++ {
				out <- newLineSplit[i]
			}
			line = newLineSplit[len(newLineSplit)-1]
		}

		if len(line) > 0 {
			out <- line
		}

	}()

	return out
}

func main() {
	var (
		listener net.Listener
		conn     net.Conn
		line     string
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

		if conn != nil {
			fmt.Print("A connection has been accepted\n")
			for line = range getLinesChannel(conn) {
				fmt.Print(line)
			}
			fmt.Print("\nThe connection has been closed.\n")
		}
	}

}
