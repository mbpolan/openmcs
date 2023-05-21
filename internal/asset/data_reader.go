package asset

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

// HasMore returns true if more data is available to be read.
func (r *DataReader) HasMore() bool {
	return r.Reader.Len() > 0
}

// VarByte reads a variable-length value, either a byte or 2-bytes, depending on the most significant byte.
func (r *DataReader) VarByte() (uint16, error) {
	msb, err := r.Byte()
	if err != nil {
		return 0, err
	}

	if msb < 0x80 {
		return uint16(msb), nil
	}

	lsb, err := r.Byte()
	if err != nil {
		return 0, err
	}

	v := (uint16(msb)<<8 | uint16(lsb)) - 0x8000
	return v, nil
}

// String reads a variable-length string.
func (r *DataReader) String() (string, error) {
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
