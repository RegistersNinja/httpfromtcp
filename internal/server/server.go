package server

import (
	"bytes"
	"fmt"
	"io"
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

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func writeHandlerError(w io.Writer, he *HandlerError) error {
	var (
		err error
	)

	if he == nil {
		return nil
	}

	err = response.WriteStatusLine(w, he.StatusCode)
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, he.Message)
	return err

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
        err            error
        defaultHeaders headers.Headers
        req            *request.Request
        buf            bytes.Buffer
        he             *HandlerError
    )

    req, err = request.RequestFromReader(conn)
    if err != nil {
        _ = writeHandlerError(conn, &HandlerError{StatusCode: response.StatusBadRequest, Message: err.Error()})
        return
    }

    he = s.handler(&buf, req)
    if he != nil {
        _ = writeHandlerError(conn, he)
        return
    }

    defaultHeaders = response.GetDefaultHeaders(buf.Len())

    err = response.WriteStatusLine(conn, response.StatusOK)
    if err != nil {
        return
    }

    err = response.WriteHeaders(conn, defaultHeaders)
    if err != nil {
        return
    }

    _, _ = io.Copy(conn, &buf)
}
