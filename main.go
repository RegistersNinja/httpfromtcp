package main


import (
	"fmt"
	"os"
)


const bytesToRead int = 8


func main() {
	var (
		messages *os.File
		err error
		bytesRead int
		buffer []byte
	)
	buffer = make([]byte, bytesToRead)


	messages, err = os.Open("messages.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open file: %v\n", err)
		os.Exit(1)
	}


	for {
		bytesRead, err = messages.Read(buffer)
		if err != nil {
			if err.Error() != "EOF" {
				fmt.Fprintf(os.Stderr, "failed to read from file: %v\n", err)
				os.Exit(1)
			}
			break
		}
		if bytesRead > 0 {
			fmt.Printf("read: %s\n", string(buffer))
		}
	}
	
}