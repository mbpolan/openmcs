package network

import (
	"bufio"
	"net"
)

// ProtocolReader parses requests sent by the client.
type ProtocolReader struct {
	reader *bufio.Reader
	buffer []byte
}

// NewReader returns a new ProtocolReader for a network connection.
func NewReader(conn net.Conn) *ProtocolReader {
	return &ProtocolReader{
		reader: bufio.NewReader(conn),
		buffer: make([]byte, 1024),
	}
}

// Byte reads a single byte from the connection.
func (r *ProtocolReader) Byte() (byte, error) {
	return r.reader.ReadByte()
}
