package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

const (
	bytesToRead int    = 8
	emptyRead   int    = 0
	ExitError   int    = 1
	channelSize int    = 1
	localAddr   string = "localhost:42069"
)

func main() {
	var (
		err         error
		udpAddr     *net.UDPAddr
		udpConn     *net.UDPConn
		stdinReader *bufio.Reader
		line        string
	)

	udpAddr, err = net.ResolveUDPAddr("udp", localAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve udp address: %v\n", err)
		os.Exit(ExitError)
	}

	udpConn, err = net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to dial udp address: %v\n", err)
		os.Exit(ExitError)
	}
	defer udpConn.Close()

	stdinReader = bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")
		line, err = stdinReader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read from buffer: %v\n", err)
			os.Exit(ExitError)
		}

		_, err = udpConn.Write([]byte(line))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to write to udp connection: %v\n", err)
			os.Exit(ExitError)
		}
	}

}
