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
	trailersState             writerState = 3
	chunkDone                 byte        = byte('0')
)

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	switch w.writerState {
	case statusLineState:
		break
	case headersState:
		return fmt.Errorf("incorrect order of response: first print status line")
	case bodyState:
		return fmt.Errorf("incorrect order of response: first print status line")
	case trailersState:
		return fmt.Errorf("incorrect order of response: response already completed")
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
	case trailersState:
		return fmt.Errorf("incorrect order of response: response already completed")
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
	case trailersState:
		return fmt.Errorf("incorrect order of response: response already completed")
	default:
		return fmt.Errorf("incorrect order of response: unknown writer state")
	}
	_, err = w.Writer.Write(body)

	return err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	switch w.writerState {
	case statusLineState:
		return 0, fmt.Errorf("incorrect order of response: first print status line")
	case headersState:
		return 0, fmt.Errorf("incorrect order of response: first print headers")
	case bodyState:
		break
	case trailersState:
		return 0, fmt.Errorf("incorrect order of response: response already completed")
	default:
		return 0, fmt.Errorf("incorrect order of response: unknown writer state")
	}

	var (
		err         error
		bytesHex    string
		bodyToWrite []byte
		bodyLength  int
	)

	bodyLength = len(p)
	bytesHex = strconv.FormatInt(int64(bodyLength), 16)

	bodyToWrite = append(bodyToWrite, []byte(bytesHex)...)
	bodyToWrite = append(bodyToWrite, []byte(crlfString)...)
	bodyToWrite = append(bodyToWrite, p...)
	bodyToWrite = append(bodyToWrite, []byte(crlfString)...)

	err = w.WriteBody(bodyToWrite)
	return bodyLength, err

}

func (w *Writer) writeChunkedBodyDone() error {
	var (
		err         error
		bodyToWrite []byte
	)

	bodyToWrite = append(bodyToWrite, chunkDone)
	bodyToWrite = append(bodyToWrite, []byte(crlfString)...)

	_, err = w.Writer.Write(bodyToWrite)
	if err != nil {
		return err
	}

	w.writerState = trailersState
	return nil
}

func (w *Writer) WriteTrailers(t headers.Headers) error {
	switch w.writerState {
	case statusLineState:
		return fmt.Errorf("incorrect order of response: first print status line")
	case headersState:
		return fmt.Errorf("incorrect order of response: first print headers")
	case bodyState:
		break
	case trailersState:
		return fmt.Errorf("incorrect order of response: response already completed")
	default:
		return fmt.Errorf("incorrect order of response: unknown writer state")
	}

	var (
		key             string
		value           string
		trailersToWrite []byte
		err             error
	)

	err = w.writeChunkedBodyDone()
	if err != nil {
		return err
	}

	for key, value = range t {
		trailersToWrite = fmt.Appendf(trailersToWrite, "%s: %s"+crlfString, key, value)
	}

	trailersToWrite = fmt.Append(trailersToWrite, crlfString)
	_, err = w.Writer.Write(trailersToWrite)
	return err
}
