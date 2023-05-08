package request

import "github.com/mbpolan/openmcs/internal/network"

const CloseInterfaceRequestHeader byte = 0x82

// CloseInterfaceRequest is sent by the client when the current interface, if any, has been dismissed.
type CloseInterfaceRequest struct {
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *CloseInterfaceRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	_, err := r.Uint8()
	if err != nil {
		return err
	}

	return nil
}
