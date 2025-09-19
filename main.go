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
	readNewLine int = 2
)

func main() {
	var (
		messages     *os.File
		err          error
		bytesRead    int
		buffer       []byte
		line         string
		newLineSplit []string
	)
	buffer = make([]byte, bytesToRead)

	messages, err = os.Open("messages.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open file: %v\n", err)
		os.Exit(ExitError)
	}

	for {
		bytesRead, err = messages.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "failed to read from file: %v\n", err)
				os.Exit(ExitError)
			}
			break
		}
		if bytesRead > emptyRead {
			line += string(buffer[:bytesRead])
		}

		newLineSplit = strings.Split(line, "\n")
		if len(newLineSplit) == readNewLine {
			fmt.Printf("read: %s\n", newLineSplit[0])
			
			line = newLineSplit[1]
			clear(newLineSplit)
		}
	}

	if len(line) > 0 {
        fmt.Printf("read: %s\n", line)
    }

}
