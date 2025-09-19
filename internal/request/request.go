package request

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/RegistersNinja/httpfromtcp/internal/headers"
)

const (
	httpVer                    string = "1.1"
	forwardSlash               string = "/"
	numberOfPartsInRequestLine int    = 3
	methodPart                 int    = 0
	pathPart                   int    = 1
	versionPart                int    = 2
	httpVersion                int    = 1
	bufferSize                 int    = 8
	initialized                int    = 0
	done                       int    = 1
	stateParsingHeaders        int    = 2
	stateParsingBody           int    = 3
	twice                      int    = 2
	ExitError                  int    = 1
	ExitSuccess                int    = 0
	newLine                    string = "\r\n"
	zeroBytesParsed            int    = 0
	clHeader                   string = "Content-Length"
	emptyStr                   string = ""
)

func getHTTPVerbs() []string {
	return []string{"GET", "HEAD", "POST", "OPTIONS", "PUT", "DELETE", "TRACE", "CONNECT"}

}

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        string
	state       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func isValidHTTPVerb(method string) bool {
	var verb string

	for _, verb = range getHTTPVerbs() {
		if method == verb {
			return true
		}
	}
	return false
}

func isValidHTTPVer(version string) bool {
	return version == httpVer
}

func (r *Request) parse(data []byte) (int, error) {
	var (
		bytesParsed     int
		reqLine         RequestLine
		err             error
		status          bool
		clValue         string
		clStr           string
		contentLength   int
		bodyBytesParsed int
	)

	switch r.state {
	case initialized:
		reqLine, bytesParsed, err = parseRequestLine(data)

		if err != nil {
			return zeroBytesParsed, err
		}
		if bytesParsed == 0 {
			return zeroBytesParsed, nil
		}

		r.RequestLine = reqLine
		r.state = stateParsingHeaders
		return bytesParsed, nil

	case stateParsingHeaders:
		bytesParsed, status, err = r.Headers.Parse(data)
		if err != nil {
			return zeroBytesParsed, err
		}
		if bytesParsed == 0 {
			return zeroBytesParsed, nil
		}
		if status {
			clValue, _ = r.Headers.Get(clHeader)
			if clValue == emptyStr {
				// No Content-Length provided: treat as zero-length body and finish.
				r.state = done
			} else {
				contentLength, err = strconv.Atoi(clValue)
				if err != nil {
					return zeroBytesParsed, fmt.Errorf("invalid Content-Length: %q", clValue)
				}
				if contentLength < 0 {
					return zeroBytesParsed, fmt.Errorf("expected non negative values for length but got %d", contentLength)
				}

				if contentLength == 0 {
					r.state = done
				} else {
					r.state = stateParsingBody
				}
			}
		}
		return bytesParsed, nil

	case stateParsingBody:
		clStr, _ = r.Headers.Get(clHeader)
		contentLength, err = strconv.Atoi(clStr)
		if err != nil {
			return zeroBytesParsed, fmt.Errorf("invalid Content-Length: %q", clStr)
		}
		if contentLength < 0 {
			return zeroBytesParsed, fmt.Errorf("expected non negative values for length but got %d", contentLength)
		}

		bodyBytesParsed = min(contentLength-len(r.Body), len(data))
		if bodyBytesParsed < 0 {
			bodyBytesParsed = 0
		}
		r.Body += string(data[:bodyBytesParsed])
		if len(r.Body) == contentLength {
			r.state = done
		}
		return bodyBytesParsed, nil
	case done:
		return zeroBytesParsed, fmt.Errorf("error: trying to read data in a done state")
	default:
		return zeroBytesParsed, fmt.Errorf("error: unknown state")
	}

}

func parseRequestLine(data []byte) (RequestLine, int, error) {
	var (
		dataNewLineSplit []string
		requestLine      string
		method           string
		target           string
		version          string
		bytesRead        int
		parts            []string
	)
	bytesRead = 0
	dataNewLineSplit = strings.Split(string(data), newLine)

	if len(dataNewLineSplit) < 2 {
		return RequestLine{}, 0, nil
	}

	requestLine = dataNewLineSplit[0]
	bytesRead = len(requestLine) + len(newLine)
	parts = strings.Split(requestLine, " ")
	if len(parts) != numberOfPartsInRequestLine {
		return RequestLine{}, bytesRead, fmt.Errorf("incorrect request line: expected %d parts, got %d", numberOfPartsInRequestLine, len(parts))
	}

	method = parts[methodPart]
	target = parts[pathPart]
	version = strings.Split(parts[versionPart], forwardSlash)[httpVersion]

	if !isValidHTTPVerb(parts[methodPart]) {
		return RequestLine{}, bytesRead, fmt.Errorf("invalid HTTP method: %s", parts[methodPart])
	}

	if !isValidHTTPVer(version) {
		return RequestLine{}, bytesRead, fmt.Errorf("invalid HTTP version, expected %s, got: %s", httpVer, parts[versionPart])
	}
	return RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   version,
	}, bytesRead, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	var (
		buf           []byte
		biggerBuf     []byte
		readToIndex   int
		bytesRead     int
		bytesParsed   int
		err           error
		parsedRequest *Request
		clStr         string
		contentLength int
	)

	buf = make([]byte, bufferSize)
	readToIndex = 0
	parsedRequest = &Request{
		state:   initialized,
		Headers: headers.NewHeaders(),
	}

	for parsedRequest.state != done {
		if readToIndex == len(buf) {
			biggerBuf = make([]byte, twice*len(buf))
			copy(biggerBuf, buf[:readToIndex])
			buf = biggerBuf
		}

		bytesRead, err = reader.Read(buf[readToIndex:])
		if err != nil {
			if err == io.EOF {
				if readToIndex > 0 {
					for {
						if parsedRequest.state == done {
							break
						}
						bytesParsed, err = parsedRequest.parse(buf[:readToIndex])
						if err != nil {
							return nil, fmt.Errorf("parse error: %s", err)
						}
						if bytesParsed == 0 {
							break
						}
						copy(buf, buf[bytesParsed:readToIndex])
						clear(buf[readToIndex-bytesParsed : readToIndex])
						readToIndex -= bytesParsed
					}
				}
				if parsedRequest.state == stateParsingBody {
					clStr, _ = parsedRequest.Headers.Get(clHeader)
					contentLength, err = strconv.Atoi(clStr)
					if err != nil || contentLength < 0 {
						return nil, fmt.Errorf("invalid Content-Length: %q", clStr)
					}
					if len(parsedRequest.Body) < contentLength {
						return nil, fmt.Errorf("unexpected EOF: body shorter than Content-Length")
					}
				}
				break
			}

			return nil, fmt.Errorf("read error: %s", err)
		}
		readToIndex += bytesRead

		for {
			bytesParsed, err = parsedRequest.parse(buf[:readToIndex])
			if err != nil {
				return nil, fmt.Errorf("parse error: %s", err)
			}
			if bytesParsed == 0 {
				break
			}
			copy(buf, buf[bytesParsed:readToIndex])
			clear(buf[readToIndex-bytesParsed : readToIndex])
			readToIndex -= bytesParsed
			if parsedRequest.state == done {
				break
			}
		}
	}

	return parsedRequest, nil
}
