package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHeaders(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("          Host: localhost:42069    \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 37, n)
	assert.False(t, done)

	// Test: Valid 2 headers with existing headers
	headers = NewHeaders()
	headers["Foo"] = "bar"
	data = []byte("Host: localhost:42069\r\nUser-Agent: tiny\r\n\r\n")

	// first header
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.False(t, done)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 23, n) // "Host: localhost:42069\r\n"

	// second header
	data = data[n:]
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.False(t, done)
	assert.Equal(t, "tiny", headers["User-Agent"])
	assert.Equal(t, 18, n) // "User-Agent: tiny\r\n"

	// existing header preserved
	assert.Equal(t, "bar", headers["Foo"])

	// Test: Valid done
	headers = NewHeaders()
	data = []byte("\r\nrest-of-body")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.True(t, done)
	assert.Equal(t, 2, n)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
