package network

import (
	"bufio"
	"net"
)

// ProtocolWriter serializes and writes responses from the server to the client.
type ProtocolWriter struct {
	*bufio.Writer
}

// NewWriter returns a new ProtocolWriter for a network connection.
func NewWriter(conn net.Conn) *ProtocolWriter {
	return &ProtocolWriter{
		Writer: bufio.NewWriter(conn),
	}
}
