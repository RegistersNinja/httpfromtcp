package headers

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type Headers map[string]string

const (
	crlf                  string = "\r\n"
	colon                 string = ":"
	space                 string = " "
	tab                   string = "\t"
	ows                   string = " \t"
	fieldNameAllowedChars string = `^[A-Za-z0-9!#$%&'*+\-.^_` + "`" + `|~]+$`
)

var validFieldNameRegex = regexp.MustCompile(fieldNameAllowedChars)

func NewHeaders() Headers {
	return map[string]string{}
}

func isValidFieldName(fieldName string) bool {
	return validFieldNameRegex.MatchString(fieldName)
}

func parseHeader(header []byte) (key string, value string, err error) {
	var (
		colonIndex      int
		byteBeforeColon byte
		namePart        []byte
		valPart         []byte
	)
	colonIndex = bytes.IndexByte(header, colon[0])
	if colonIndex < 0 {
		return key, value, fmt.Errorf("error: expected a key:value pair but found %q", string(header))
	}

	if colonIndex > 0 {
		if colonIndex == 0 {
			return key, value, fmt.Errorf("error: field-name must be at least one character long, got: %q", string(header))

		}
		byteBeforeColon = header[colonIndex-1]
		if byteBeforeColon == space[0] || byteBeforeColon == tab[0] {
			return key, value, fmt.Errorf("error: field name must not have whitespace before colon: %q", string(header[:colonIndex]))
		}
	}

	namePart = bytes.TrimLeft(header[:colonIndex], ows)
	if !isValidFieldName(string(namePart)) {
		return key, value, fmt.Errorf("error: invalid field-name %q", string(namePart))
	}

	valPart = bytes.TrimSpace(header[colonIndex+1:])

	return strings.ToLower(string(namePart)), string(valPart), nil
}

func (h Headers) Parse(data []byte) (bytesConsumed int, done bool, err error) {
	var (
		crlfSplit        [][]byte
		header           []byte
		fieldName        string
		fieldValue       string
		isFieldNameInMap bool
	)

	crlfSplit = bytes.SplitN(data, []byte(crlf), 2)
	if len(crlfSplit) < 2 {
		return bytesConsumed, done, err
	}
	header = crlfSplit[0]

	if len(header) == 0 {
		done = true
		bytesConsumed = len(crlf)
		return bytesConsumed, done, err
	}

	fieldName, fieldValue, err = parseHeader(header)
	if err != nil {
		return bytesConsumed, done, err
	}

	_, isFieldNameInMap = h[fieldName]
	if isFieldNameInMap {
		h[fieldName] += (", " + fieldValue)
	} else {
		h[fieldName] = fieldValue
	}

	bytesConsumed = len(header) + len(crlf)
	return bytesConsumed, done, nil
}
