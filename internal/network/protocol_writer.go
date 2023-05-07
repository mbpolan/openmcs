package network

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

// ProtocolWriter serializes and writes response from the server to the client.
type ProtocolWriter struct {
	buffer *bytes.Buffer
	writer io.Writer
}

// NewWriter returns a new ProtocolWriter for a network connection.
func NewWriter(conn net.Conn) *ProtocolWriter {
	return &ProtocolWriter{
		writer: conn,
	}
}

// NewBufferedWriter returns a new ProtocolWriter that writes to an internal buffer instead of to a network connection.
func NewBufferedWriter() *ProtocolWriter {
	buffer := &bytes.Buffer{}

	return &ProtocolWriter{
		buffer: buffer,
		writer: buffer,
	}
}

// Buffer returns the buffer representing the backing storage for a buffered ProtocolWriter. If the writer is not
// buffered, an error will be returned instead.
func (w *ProtocolWriter) Buffer() (*bytes.Buffer, error) {
	if w.buffer == nil {
		return nil, fmt.Errorf("not a buffered writer")
	}

	return w.buffer, nil
}

// Write attempts to write the slice of bytes, returning how many bytes were written and an error if the entire slice
// could not be written.
func (w *ProtocolWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

// WriteUint8 writes a single, unsigned byte.
func (w *ProtocolWriter) WriteUint8(n uint8) error {
	_, err := w.writer.Write([]byte{n})
	if err != nil {
		return err
	}

	return nil
}

// WriteVarByte writes either a single, unsigned byte or two unsigned bytes depending on the value.
func (w *ProtocolWriter) WriteVarByte(n uint16) error {
	if n < 0x80 {
		return w.WriteUint8(uint8(n))
	}

	return w.WriteUint16(n + 0x8000)
}

// WriteUint16Alt2 writes an unsigned, 16-bit (short) integer using the alternative format.
func (w *ProtocolWriter) WriteUint16Alt2(n uint16) error {
	err := w.WriteUint8(byte(n + 0x80))
	if err != nil {
		return err
	}

	err = w.WriteUint8(byte(n >> 8))
	if err != nil {
		return err
	}

	return nil
}

// WriteUint16Alt writes an unsigned, 16-bit (short) integer using the alternative format.
func (w *ProtocolWriter) WriteUint16Alt(n uint16) error {
	err := w.WriteUint8(byte(n >> 8))
	if err != nil {
		return err
	}

	err = w.WriteUint8(byte(n - 128))
	if err != nil {
		return err
	}

	return nil
}

// WriteUint16 writes an unsigned, 16-bit (short) integer.
func (w *ProtocolWriter) WriteUint16(n uint16) error {
	err := w.WriteUint8(byte(n >> 8))
	if err != nil {
		return err
	}

	err = w.WriteUint8(byte(n))
	if err != nil {
		return err
	}

	return nil
}

// WriteUint16LE writes an unsigned, 16-bit (short) integer in little-endian format.
func (w *ProtocolWriter) WriteUint16LE(n uint16) error {
	err := w.WriteUint8(byte(n))
	if err != nil {
		return err
	}

	err = w.WriteUint8(byte(n >> 8))
	if err != nil {
		return err
	}

	return nil
}

// WriteUint32 writes an unsigned, 32-bit integer.
func (w *ProtocolWriter) WriteUint32(n uint32) error {
	err := w.WriteUint8(byte(n >> 24))
	if err != nil {
		return err
	}

	err = w.WriteUint8(byte(n >> 16))
	if err != nil {
		return err
	}

	err = w.WriteUint8(byte(n >> 8))
	if err != nil {
		return err
	}

	err = w.WriteUint8(byte(n))
	if err != nil {
		return err
	}

	return nil
}

// WriteUint64 writes an unsigned, 64-bit (long) integer.
func (w *ProtocolWriter) WriteUint64(n uint64) error {
	err := w.WriteUint32(uint32(n >> 32))
	if err != nil {
		return err
	}

	err = w.WriteUint32(uint32(n))
	if err != nil {
		return err
	}

	return nil
}

// WriteString writes a variable-length string.
func (w *ProtocolWriter) WriteString(s string) error {
	// write each character as a single byte
	for _, ch := range s {
		err := w.WriteUint8(byte(ch))
		if err != nil {
			return err
		}
	}

	// write a byte to indicate the end of the string
	err := w.WriteUint8(0x0A)
	if err != nil {
		return err
	}

	return nil
}
