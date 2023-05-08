package request

import "github.com/mbpolan/openmcs/internal/network"

const RegionLoadedRequestHeader byte = 0x79

// RegionLoadedRequest is sent by the client when the player entered a new region.
type RegionLoadedRequest struct {
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *RegionLoadedRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	_, err := r.Uint8()
	if err != nil {
		return err
	}

	return nil
}
