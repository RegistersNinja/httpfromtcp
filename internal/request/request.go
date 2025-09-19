package request

import (
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	httpVer                    string = "1.1"
	numberOfPartsInRequestLine int    = 3
	methodPart                 int    = 0
	pathPart                   int    = 1
	versionPart                int    = 2
	httpVersion                int    = 1
)

func getHTTPVerbs() []string {
	return []string{"GET", "HEAD", "POST", "OPTIONS", "PUT", "DELETE", "TRACE", "CONNECT"}

}

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func isValidHTTPVerb(method string) bool {
	for _, verb := range getHTTPVerbs() {
		if method == verb {
			return true
		}
	}
	return false
}

func isValidHTTPVer(version string) bool {
	return version == httpVer
}

func parseRequestLine(requestLine string) (RequestLine, error) {

	var (
		method  string
		target  string
		version string
	)
	parts := strings.SplitN(requestLine, " ", numberOfPartsInRequestLine)
	if len(parts) != numberOfPartsInRequestLine {
		return RequestLine{}, fmt.Errorf("incorrect request line: expected %d parts, got %d", numberOfPartsInRequestLine, len(parts))
	}

	method = parts[methodPart]
	target = parts[pathPart]
	version = strings.Split(parts[versionPart], "/")[httpVersion]

	if !isValidHTTPVerb(parts[methodPart]) {
		return RequestLine{}, fmt.Errorf("invalid HTTP method: %s", parts[methodPart])
	}

	if !isValidHTTPVer(version) {
		return RequestLine{}, fmt.Errorf("invalid HTTP version, expected %s, got: %s", httpVer, parts[versionPart])
	}
	return RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   version,
	}, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	var (
		requestRead       []byte
		requestReadString string
		requestLine       string
		err               error
		parsedLine        RequestLine
		parsedRequest     *Request
	)

	requestRead, err = io.ReadAll(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read: %v\n", err)
		return nil, err
	}

	requestReadString = string(requestRead)

	requestLine = strings.Split(requestReadString, "\r\n")[0]

	parsedLine, err = parseRequestLine(requestLine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse request line: %v\n", err)
		return nil, err
	}

	parsedRequest = &Request{
		RequestLine: parsedLine,
	}

	return parsedRequest, nil
}
