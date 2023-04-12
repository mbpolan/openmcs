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

// WriteUint16 writes an unsigned, 16-bit (short) integer.
func (w *ProtocolWriter) WriteUint16(n uint16) error {
	err := w.WriteByte(byte(n >> 8))
	if err != nil {
		return err
	}

	err = w.WriteByte(byte(n))
	if err != nil {
		return err
	}

	return nil
}

// WriteUint64 writes an unsigned, 64-bit (long) integer.
func (w *ProtocolWriter) WriteUint64(n uint64) error {
	err := w.WriteByte(byte(n >> 56))
	if err != nil {
		return err
	}

	err = w.WriteByte(byte(n >> 48))
	if err != nil {
		return err
	}

	err = w.WriteByte(byte(n >> 40))
	if err != nil {
		return err
	}

	err = w.WriteByte(byte(n >> 32))
	if err != nil {
		return err
	}

	err = w.WriteByte(byte(n >> 24))
	if err != nil {
		return err
	}

	err = w.WriteByte(byte(n >> 16))
	if err != nil {
		return err
	}

	err = w.WriteByte(byte(n >> 8))
	if err != nil {
		return err
	}

	err = w.WriteByte(byte(n))
	if err != nil {
		return err
	}

	return nil
}
