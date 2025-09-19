package response

import (
	"io"
	"strconv"

	"github.com/RegistersNinja/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
	httpVerString             string     = "HTTP/1.1"
	spaceString               string     = " "
	lineOKString              string     = "200 OK"
	lineBRString              string     = "400 Bad Request"
	lineSEString              string     = "500 Internal Server Error"
	clString                  string     = "Content-Length"
	conString                 string     = "Connection"
	conValString              string     = "close"
	ctString                  string     = "Content-Type"
	ctValString               string     = "text/plain"
	crlfString                string     = "\r\n"
	colonString               string     = ":"
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var err error
	switch statusCode {
	case StatusOK:
		_, err = w.Write([]byte(httpVerString + spaceString + lineOKString + crlfString))

	case StatusBadRequest:
		_, err = w.Write([]byte(httpVerString + spaceString + lineBRString + crlfString))

	case StatusInternalServerError:
		_, err = w.Write([]byte(httpVerString + spaceString + lineSEString + crlfString))
	}

	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	var h headers.Headers = headers.NewHeaders()

	h[clString] = strconv.Itoa(contentLen)
	h[conString] = conValString
	h[ctString] = ctValString

	return h
}

func serializeHeadersMap(h headers.Headers) []string {
	var (
		responseStrings []string
		keys            []string
		key             string
		index           int
	)

	keys = make([]string, 0, len(h))
	for key = range h {
		keys = append(keys, key)
	}

	responseStrings = make([]string, len(keys))
	for index, key = range keys {
		responseStrings[index] = key + colonString + spaceString + h[key] + crlfString
	}

	return responseStrings
}

func WriteHeaders(w io.Writer, h headers.Headers) error {
	var (
		lines []string
		line  string
		err   error
	)

	lines = serializeHeadersMap(h)
	for _, line = range lines {
		if _, err = io.WriteString(w, line); err != nil {
			return err
		}
	}
	_, err = io.WriteString(w, crlfString)
	return err
}
