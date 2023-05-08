package request

import "github.com/mbpolan/openmcs/internal/network"

const FocusRequestHeader byte = 0x03

// FocusChangeRequest is sent by the client when the client window loses or acquires focus.
type FocusChangeRequest struct {
	Focused bool
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *FocusChangeRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	_, err := r.Uint8()
	if err != nil {
		return err
	}

	// read a single byte, indicating if the client window is current focused or not
	b, err := r.Uint8()
	if err != nil {
		return err
	}

	p.Focused = b == 0x01
	return nil
}
