package request

import "github.com/mbpolan/openmcs/internal/network"

const RegionChangeRequestHeader byte = 0xD2

// RegionChangeRequest is sent by the client when the player enters a new map region.
type RegionChangeRequest struct {
	Flag int
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *RegionChangeRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	_, err := r.Uint8()
	if err != nil {
		return err
	}

	// read 4 bytes containing some unknown value
	flag, err := r.Uint32()
	if err != nil {
		return err
	}

	p.Flag = int(flag)
	return nil
}
