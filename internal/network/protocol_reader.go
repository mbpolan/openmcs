package network

import (
	"bufio"
	"net"
)

// ProtocolReader parses requests sent by the client.
type ProtocolReader struct {
	*bufio.Reader
	buffer []byte
}

// NewReader returns a new ProtocolReader for a network connection.
func NewReader(conn net.Conn) *ProtocolReader {
	return &ProtocolReader{
		Reader: bufio.NewReader(conn),
		buffer: make([]byte, 1024),
	}
}

// Skip reads exactly n bytes and discards them.
func (r *ProtocolReader) Skip(n int) error {
	for i := 0; i < n; i++ {
		_, err := r.Byte()
		if err != nil {
			return err
		}
	}

	return nil
}

// Byte reads a single byte from the connection.
func (r *ProtocolReader) Byte() (byte, error) {
	return r.Reader.ReadByte()
}

// Uint16 reads an unsigned, 16-bit integer in big-endian format.
func (r *ProtocolReader) Uint16() (uint16, error) {
	i1, err := r.Byte()
	if err != nil {
		return 0, err
	}

	i2, err := r.Byte()
	if err != nil {
		return 0, err
	}

	return (uint16(i1) << 8) | uint16(i2), nil
}

// Uint32 reads an unsigned, 32-bit integer in big-endian format.
func (r *ProtocolReader) Uint32() (uint32, error) {
	i1, err := r.Byte()
	if err != nil {
		return 0, err
	}

	i2, err := r.Byte()
	if err != nil {
		return 0, err
	}

	i3, err := r.Byte()
	if err != nil {
		return 0, err
	}

	i4, err := r.Byte()
	if err != nil {
		return 0, err
	}

	return (uint32(i1) << 24) | (uint32(i2) << 16) | (uint32(i3) << 8) | uint32(i4), nil
}

// String reads a variable-length string.
func (r *ProtocolReader) String() (string, error) {
	var str []byte

	for {
		ch, err := r.Byte()
		if err != nil {
			return "", err
		}

		// 0x0A indicates end of string
		if ch == 0x0A {
			break
		} else {
			str = append(str, ch)
		}
	}

	return string(str), nil
}
