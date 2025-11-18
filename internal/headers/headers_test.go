package headers

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)


func TestHeaderParse(t *testing.T) {
    // Test: Valid single header
    headers := NewHeaders()
    data := []byte("Host: localhost:42069\r\n\r\n")
    n, done, err := headers.Parse(data)
    require.NoError(t, err)
    require.NotNil(t, headers)
    assert.Equal(t, "localhost:42069", headers["host"])
    assert.Equal(t, 23, n)
    assert.False(t, done)

    // Test: Valid single header with extra whitespace
    headers = NewHeaders()
    data = []byte("        Host:          localhost:42069         \r\n\r\n")
    n, done, err = headers.Parse(data)
    require.NoError(t, err)
    require.NotNil(t, headers)
    assert.Equal(t, "localhost:42069", headers["host"])
    assert.Equal(t, 49, n)
    assert.False(t, done)

    // Test: valid done
    headers = NewHeaders()
    data = []byte("\r\n")
    n, done, err = headers.Parse(data)
    require.NoError(t, err)
    require.NotNil(t, headers)
    assert.Equal(t, 2, n)
    assert.True(t, done)

    // Test: valid 2 headers with existing headers
    headers = NewHeaders()
    headers["host"] = "localhost:42069"
    data = []byte("Content-Type: text/html\r\n\r\n")
    n, done, err = headers.Parse(data)
    require.NoError(t, err)
    require.NotNil(t, headers)
    require.Equal(t, "localhost:42069", headers["host"])
    require.Equal(t, "text/html", headers["content-type"])
    assert.Equal(t, 25, n)
    assert.False(t, done)

    // Test: Invalid spacing header
    headers = NewHeaders()
    data = []byte("       Host : localhost:42069       \r\n\r\n")
    n, done, err = headers.Parse(data)
    require.Error(t, err)
    assert.Equal(t, 0, n)
    assert.False(t, done)

    // Test: Field value with multiple elements
    headers = NewHeaders()
    data = []byte("Example-Dates: \"Sat, 04 May 1996\", \"Wed, 14 Sep 2005\"\r\n\r\n")
    n, done, err = headers.Parse(data)
    require.NoError(t, err)
    require.NotNil(t, headers)
    require.Equal(t, "\"Sat, 04 May 1996\", \"Wed, 14 Sep 2005\"", headers["example-dates"])
    assert.Equal(t, 55, n)
    assert.False(t, done)

    // Test: Field name includes invalid character
    headers = NewHeaders()
    data = []byte("Höst: localhost:42069\r\n\r\n")
    n, done, err = headers.Parse(data)
    require.Error(t, err)
    assert.Equal(t, 0, n)
    assert.False(t, done)

    // Test: empty field name
    headers = NewHeaders()
    data = []byte(": localhost:42069\r\n\r\n")
    n, done, err = headers.Parse(data)
    require.Error(t, err)
    assert.Equal(t, 0, n)
    assert.False(t, done)

    // Test: multiple matching field names
    headers = NewHeaders()
    headers["set-person"] = "lane-loves-go"
    data = []byte("Set-Person: prime-loves-zig\r\n\r\n")
    n, done, err = headers.Parse(data)
    require.NoError(t, err)
    require.NotNil(t, headers)
    require.Equal(t, "lane-loves-go, prime-loves-zig", headers["set-person"])
    assert.Equal(t, 29, n)
    assert.False(t, done)
}
