package request

import "github.com/mbpolan/openmcs/internal/network"

const InterfaceActionRequestHeader byte = 0xB9

// InterfaceActionRequest is sent by the client when the player interacts with an interface.
type InterfaceActionRequest struct {
	Action int
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *InterfaceActionRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	_, err := r.Uint8()
	if err != nil {
		return err
	}

	// read 2 bytes containing the interface action id
	action, err := r.Uint16()
	if err != nil {
		return err
	}

	p.Action = int(action)
	return nil
}
