package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	bytesToRead int = 8
	emptyRead   int = 0
	ExitError   int = 1
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)

	go func() {
		defer f.Close()
		defer close(out)
		
		var (
			buffer       []byte
			line         string
			newLineSplit []string
		)

		buffer = make([]byte, bytesToRead)
		for {
			bytesRead, err := f.Read(buffer)
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
		messages *os.File
		err      error
	)

	messages, err = os.Open("messages.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open file: %v\n", err)
		os.Exit(ExitError)
	}

	for line := range getLinesChannel(messages) {
		fmt.Printf("read: %s\n", line)
	}

}
