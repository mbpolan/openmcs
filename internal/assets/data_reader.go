package assets

import (
	"bytes"
)

type DataReader struct {
	*bytes.Reader
}

func NewDataReader(b []byte) *DataReader {
	return &DataReader{Reader: bytes.NewReader(b)}
}

// Byte reads a single byte.
func (r *DataReader) Byte() (byte, error) {
	return r.Reader.ReadByte()
}

// Uint16 reads an unsigned, 16-bit integer in big-endian format.
func (r *DataReader) Uint16() (uint16, error) {
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

// Uint24 reads an unsigned, 24-bit integer in big-endian format.
func (r *DataReader) Uint24() (uint32, error) {
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

	return (uint32(i1) << 16) | (uint32(i2) << 8) | uint32(i3), nil
}

// Uint32 reads an unsigned, 32-bit integer in big-endian format.
func (r *DataReader) Uint32() (uint32, error) {
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
