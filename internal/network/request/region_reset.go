package request

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/network"
)

const RegionResetRequestHeader byte = 0x96

// RegionResetRequest is sent by the client to report it has loaded a certain number of map regions.
type RegionResetRequest struct {
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *RegionResetRequest) Read(r *network.ProtocolReader) error {
	// read packet header
	b, err := r.Uint8()
	if err != nil {
		return err
	}

	if b != RegionResetRequestHeader {
		return fmt.Errorf("unexpected packet header")
	}

	return nil
}