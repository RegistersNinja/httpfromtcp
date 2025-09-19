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
	assert.Equal(t, "localhost:42069", headers["host"]) // Map key is lowercase
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
	assert.Equal(t, "localhost:42069", headers["host"]) // Map key is lowercase
	assert.Equal(t, 37, n)
	assert.False(t, done)

	// Test: Valid 2 headers with existing headers
	headers = NewHeaders()
	headers["foo"] = "bar" // Existing header key is lowercase
	data = []byte("Host: localhost:42069\r\nUser-Agent: tiny\r\n\r\n")

	// first header
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.False(t, done)
	assert.Equal(t, "localhost:42069", headers["host"]) // Map key is lowercase
	assert.Equal(t, 23, n)                              // "Host: localhost:42069\r\n"

	// second header
	data = data[n:]
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.False(t, done)
	assert.Equal(t, "tiny", headers["user-agent"]) // Map key is lowercase
	assert.Equal(t, 18, n)                         // "User-Agent: tiny\r\n"

	// existing header preserved
	assert.Equal(t, "bar", headers["foo"]) // Existing header key is lowercase

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

	// Test: Header key with capital letters
	headers = NewHeaders()
	data = []byte("X-Custom-Header: value\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "value", headers["x-custom-header"]) // Map key is lowercase
	assert.Equal(t, 24, n)
	assert.False(t, done)

	// Test: Invalid character in header key
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Multiple values for the same header key
    headers = NewHeaders()
    headers["set-person"] = "lane-loves-go" // Existing header key is lowercase
    data = []byte("Set-Person: prime-loves-zig\r\nSet-Person: tj-loves-ocaml\r\n\r\n")

    // Parse first header
    n, done, err = headers.Parse(data)
    require.NoError(t, err)
    assert.False(t, done)
    assert.Equal(t, "lane-loves-go, prime-loves-zig", headers["set-person"])
    assert.Equal(t, 29, n) // "Set-Person: prime-loves-zig\r\n"

    // Parse second header
    data = data[n:]
    n, done, err = headers.Parse(data)
    require.NoError(t, err)
    assert.False(t, done)
    assert.Equal(t, "lane-loves-go, prime-loves-zig, tj-loves-ocaml", headers["set-person"])
    assert.Equal(t, 28, n) // "Set-Person: tj-loves-ocaml\r\n"
}
