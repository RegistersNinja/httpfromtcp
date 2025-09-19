package server

import (
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"github.com/RegistersNinja/httpfromtcp/internal/headers"
	"github.com/RegistersNinja/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	var (
		listener net.Listener
		err      error
		s        *Server
	)

	listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	s = &Server{listener: listener}
	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	if s == nil {
		return nil
	}
	s.closed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) listen() {
	var (
		conn net.Conn
		err  error
	)

	for {
		conn, err = s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	var (
		err            error
		body           string
		defaultHeaders headers.Headers
	)

	body = "Hello World!"

	if err = response.WriteStatusLine(conn, response.StatusOK); err != nil {
		return
	}

	defaultHeaders = response.GetDefaultHeaders(len(body))
	if err := response.WriteHeaders(conn, defaultHeaders); err != nil {
		return
	}

	if _, err = io.WriteString(conn, body); err != nil {
		return
	}
}
