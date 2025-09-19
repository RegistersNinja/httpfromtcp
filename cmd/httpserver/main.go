package main

import (
	"crypto/sha256"
	"encoding/hex"
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

const (
	port int = 42069

	proxyPath string = "https://httpbin.org/"

	headerContentType   string = "Content-Type"
	headerConnection    string = "Connection"
	headerTransferEnc   string = "Transfer-Encoding"
	headerTrailer       string = "Trailer"
	headerContentLength string = "Content-Length"

	trailerContentSHA    string = "X-Content-SHA256"
	trailerContentLength string = "X-Content-Length"

	trailerAnnouncement string = trailerContentSHA + ", " + trailerContentLength

	connectionClose         string = "close"
	transferEncodingChunked string = "chunked"
	defaultProxyBufferSize  int    = 1024
)

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
			status           response.StatusCode
			hdrs             headers.Headers
			body             []byte
			path             string
			url              string
			resp             *http.Response
			ct               string
			buf              []byte
			n                int
			rerr             error
			handled          bool
			requestTarget    string
			trailer          headers.Headers
			fullBody         []byte
			hash             [sha256.Size]byte
			hashString       string
			contentLenString string
		)

		requestTarget = req.RequestLine.RequestTarget
		switch {
		case strings.HasPrefix(requestTarget, "/httpbin/"):
			path = strings.TrimPrefix(requestTarget, "/httpbin/")
			url = proxyPath + path

			resp, err = http.Get(url)
			if err != nil {
				status = response.StatusInternalServerError
				body = respond500()
				hdrs = response.GetDefaultHeaders(len(body))
				hdrs.OverrideHeaderValue(headerContentLength, strconv.Itoa(len(body)))
				_ = w.WriteStatusLine(status)
				_ = w.WriteHeaders(hdrs)
				_ = w.WriteBody(body)
				return
			}
			defer resp.Body.Close()

			_ = w.WriteStatusLine(response.StatusOK)
			hdrs = headers.NewHeaders()
			ct = resp.Header.Get(headerContentType)
			if ct != "" {
				hdrs.OverrideHeaderValue(headerContentType, ct)
			}
			hdrs.OverrideHeaderValue(headerConnection, connectionClose)
			hdrs.OverrideHeaderValue(headerTransferEnc, transferEncodingChunked)
			hdrs.OverrideHeaderValue(headerTrailer, trailerAnnouncement)
			_ = w.WriteHeaders(hdrs)

			buf = make([]byte, defaultProxyBufferSize)
			for {
				n, rerr = resp.Body.Read(buf)
				if n > 0 {
					fullBody = append(fullBody, buf[:n]...)
					_, _ = w.WriteChunkedBody(buf[:n])
				}
				if rerr == io.EOF {
					hash = sha256.Sum256(fullBody)
					hashString = hex.EncodeToString(hash[:])
					contentLenString = strconv.Itoa(len(fullBody))
					trailer = headers.NewHeaders()
					trailer.OverrideHeaderValue(trailerContentSHA, hashString)
					trailer.OverrideHeaderValue(trailerContentLength, contentLenString)
					_ = w.WriteTrailers(trailer)
					break
				}
				if rerr != nil {
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

		if !handled {
			hdrs = response.GetDefaultHeaders(len(body))
			hdrs.OverrideHeaderValue(headerContentLength, strconv.Itoa(len(body)))
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
