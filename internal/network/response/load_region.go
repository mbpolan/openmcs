package response

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
)

const LoadRegionResponseHeader byte = 0x49

// LoadRegionResponse instructs the client to load a specific region of the world map.
type LoadRegionResponse struct {
	region model.Vector2D
}

// NewLoadRegionResponse creates a new map region load response. The region should be specified in region origin
// coordinates.
func NewLoadRegionResponse(region model.Vector2D) *LoadRegionResponse {
	return &LoadRegionResponse{
		region: region,
	}
}

// Write writes the contents of the message to a stream.
func (p *LoadRegionResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(LoadRegionResponseHeader)
	if err != nil {
		return err
	}

	// write x and y coordinates
	err = w.WriteUint16Alt(uint16(p.region.X))
	if err != nil {
		return err
	}

	err = w.WriteUint16(uint16(p.region.Y))
	if err != nil {
		return err
	}

	return nil
}
