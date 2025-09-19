package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/RegistersNinja/httpfromtcp/internal/headers"
)

type StatusCode int
type writerState int

type Writer struct {
	Writer      io.Writer
	writerState writerState
}

const (
	StatusOK                  StatusCode  = 200
	StatusBadRequest          StatusCode  = 400
	StatusInternalServerError StatusCode  = 500
	httpVerString             string      = "HTTP/1.1"
	spaceString               string      = " "
	lineOKString              string      = "200 OK"
	lineBRString              string      = "400 Bad Request"
	lineSEString              string      = "500 Internal Server Error"
	clString                  string      = "Content-Length"
	conString                 string      = "Connection"
	conValString              string      = "close"
	ctString                  string      = "Content-Type"
	ctValString               string      = "text/plain"
	crlfString                string      = "\r\n"
	statusLineState           writerState = 0
	headersState              writerState = 1
	bodyState                 writerState = 2
)

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
    switch w.writerState {
    case statusLineState:
        break
	case headersState:
		return fmt.Errorf("incorrect order of response: first print status line")
	case bodyState:
		return fmt.Errorf("incorrect order of response: first print status line")
	default:
		return fmt.Errorf("incorrect order of response: unknown writer state")
	}

    var (
        err        error
        statusLine string
    )

    statusLine = httpVerString + spaceString
    switch statusCode {
    case StatusOK:
        statusLine += lineOKString
    case StatusBadRequest:
        statusLine += lineBRString
    case StatusInternalServerError:
        statusLine += lineSEString
    }
    statusLine += crlfString

    _, err = w.Writer.Write([]byte(statusLine))
    w.writerState = headersState
    return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	var h headers.Headers = headers.NewHeaders()

	h[clString] = strconv.Itoa(contentLen)
	h[conString] = conValString
	h[ctString] = ctValString

	return h
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	switch w.writerState {
	case statusLineState:
		return fmt.Errorf("incorrect order of response: first print status line")
	case headersState:
		break
	case bodyState:
		return fmt.Errorf("incorrect order of response: first print headers")
	default:
		return fmt.Errorf("incorrect order of response: unknown writer state")
	}
	var (
		key            string
		value          string
		headersToWrite []byte
		err            error
	)

	for key, value = range h {
		headersToWrite = fmt.Appendf(headersToWrite, "%s: %s\r\n", key, value)
	}

	headersToWrite = fmt.Append(headersToWrite, crlfString)
	_, err = w.Writer.Write(headersToWrite)
	w.writerState = bodyState
	return err
}

func (w *Writer) WriteBody(body []byte) (err error) {
	switch w.writerState {
	case statusLineState:
		return fmt.Errorf("incorrect order of response: first print status line")
	case headersState:
		return fmt.Errorf("incorrect order of response: first print headers")
	case bodyState:
		break
	default:
		return fmt.Errorf("incorrect order of response: unknown writer state")
	}
	_, err = w.Writer.Write(body)
	
	return err
}
