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


	bytesRead, err = messages.Read(buffer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read from file: %v\n", err)
		os.Exit(1)
	}

	for(bytesRead == 8) {
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read from file: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("read: %s\n", string(buffer))
		clear(buffer)
		bytesRead, err = messages.Read(buffer)
	}
	
	if err == nil {
		fmt.Printf("read: %s\n", string(buffer))
	} else {
			fmt.Fprintf(os.Stderr, "failed to read from file: %v\n", err)
			os.Exit(1)
	}
}