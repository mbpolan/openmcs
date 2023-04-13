package requests

import "github.com/mbpolan/openmcs/internal/network"

const RegionChangeRequestHeader byte = 0xD2

// RegionChangeRequest is sent by the client when the player enters a new map region.
type RegionChangeRequest struct {
}

func ReadRegionChangeRequest(r *network.ProtocolReader) (*RegionChangeRequest, error) {
	// read 4 bytes containing some unknown value
	_, err := r.Uint32()
	if err != nil {
		return nil, err
	}

	return &RegionChangeRequest{}, nil
}
