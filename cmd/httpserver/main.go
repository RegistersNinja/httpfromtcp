package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/RegistersNinja/httpfromtcp/internal/headers"
	"github.com/RegistersNinja/httpfromtcp/internal/request"
	"github.com/RegistersNinja/httpfromtcp/internal/response"
	"github.com/RegistersNinja/httpfromtcp/internal/server"
)

const port = 42069

func respond400() []byte {
	return []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
}

func respond500() []byte {
	return []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
}

func respond200() []byte {
	return []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
}

func main() {
	var (
		handler server.Handler
		srv     *server.Server
		err     error
		sigChan chan os.Signal
	)

		handler = func(w *response.Writer, req *request.Request) {
			var (
				status response.StatusCode
				hdrs   headers.Headers
				body   []byte
				path   string
				url    string
				resp   *http.Response
				ct     string
				buf    []byte
				n      int
				rerr   error
				handled bool
				requestTarget string
			)

		const proxyPath string = "https://httpbin.org/"
		requestTarget = req.RequestLine.RequestTarget  
		switch {
		case strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/"):
			path = strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
			url = proxyPath + path

			resp, err = http.Get(url)
			if err != nil {
				status = response.StatusInternalServerError
				body = respond500()
				hdrs = response.GetDefaultHeaders(len(body))
				hdrs.OverrideHeaderValue("Content-Length", strconv.Itoa(len(body)))
				w.WriteStatusLine(status)
				w.WriteHeaders(hdrs)
				w.WriteBody(body)
				return
			}
			defer resp.Body.Close()

			w.WriteStatusLine(response.StatusOK)
			hdrs = headers.NewHeaders()
			ct = resp.Header.Get("Content-Type")
			if ct != "" {
				hdrs.OverrideHeaderValue("Content-Type", ct)
			}
			hdrs.OverrideHeaderValue("Connection", "close")
			hdrs.OverrideHeaderValue("Transfer-Encoding", "chunked")
			w.WriteHeaders(hdrs)

			buf = make([]byte, 1024)
			for {
				n, rerr = resp.Body.Read(buf)
				if n > 0 {
					_, _ = w.WriteChunkedBody(buf[:n])
				}
				if rerr == io.EOF {
					_, _ = w.WriteChunkedBodyDone()
					break
				}
				if rerr != nil {
					// Abort on read error
					break
				}
				}
				handled = true

		case requestTarget == "/yourproblem":
			status = response.StatusBadRequest
			body = respond400()

		case requestTarget == "/myproblem":
			status = response.StatusInternalServerError
			body = respond500()

		default:
			status = response.StatusOK
			body = respond200()
		}

			// Non-proxy: fixed-length response
			if !handled {
				hdrs = response.GetDefaultHeaders(len(body))
				hdrs.OverrideHeaderValue("Content-Length", strconv.Itoa(len(body)))
				_ = w.WriteStatusLine(status)
				_ = w.WriteHeaders(hdrs)
				_ = w.WriteBody(body)
			}
		}

	srv, err = server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()
	log.Println("Server started on port", port)

	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
