package requests

import "github.com/mbpolan/openmcs/internal/network"

const RegionLoadedRequestHeader byte = 0x79

// RegionLoadedRequest is sent by the client when it finished loading a map region.
type RegionLoadedRequest struct {
}

func ReadRegionLoadedRequest(r *network.ProtocolReader) (*RegionLoadedRequest, error) {
	// no payload
	return &RegionLoadedRequest{}, nil
}
