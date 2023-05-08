package request

import "github.com/mbpolan/openmcs/internal/network"

const KeepAliveRequestHeader byte = 0x00

// KeepAliveRequest is sent by the client to maintain connectivity.
type KeepAliveRequest struct {
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *KeepAliveRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	_, err := r.Uint8()
	if err != nil {
		return err
	}

	return nil
}
