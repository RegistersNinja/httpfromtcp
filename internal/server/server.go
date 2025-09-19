package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/RegistersNinja/httpfromtcp/internal/headers"
	"github.com/RegistersNinja/httpfromtcp/internal/request"
	"github.com/RegistersNinja/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handler  Handler
}

type Handler func(w *response.Writer, req *request.Request)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func Serve(port int, handleFunc Handler) (*Server, error) {
	var (
		listener net.Listener
		err      error
		s        *Server
	)

	listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	s = &Server{listener: listener, handler: handleFunc}
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
		rpWriter *response.Writer
		req      *request.Request
		err      error
		body     []byte
		hdrs     headers.Headers
	)

	rpWriter = &response.Writer{Writer: conn}

	req, err = request.RequestFromReader(conn)
	if err != nil {
		rpWriter.WriteStatusLine(response.StatusBadRequest)
		body = []byte(err.Error())
		hdrs = response.GetDefaultHeaders(len(body))
		rpWriter.WriteHeaders(hdrs)
		rpWriter.WriteBody(body)
		return
	}

	s.handler(rpWriter, req)
}
