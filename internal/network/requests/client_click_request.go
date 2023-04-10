package requests

import "github.com/mbpolan/openmcs/internal/network"

const ClientClickRequestHeader byte = 0xF1

// ClientClickRequest is sent by the client when the player clicks on an area of the client window.
type ClientClickRequest struct {
	TimeSince   uint32
	RightClick  bool
	PixelOffset uint32
}

// ReadClientClickRequest parses the packet from the connection stream.
func ReadClientClickRequest(r *network.ProtocolReader) (*ClientClickRequest, error) {
	// read a single integer that contains packed data
	v, err := r.Uint32()
	if err != nil {
		return nil, err
	}

	// time since last click is contained in the high 12 bits
	timeSince := v & 0xFFF

	// left or right click is in the 13th bit
	rightClick := v & 0x80000

	// pixel offset is in the remaining bits
	pixelOffset := v & 0x7FFFF

	return &ClientClickRequest{
		TimeSince:   timeSince,
		RightClick:  rightClick == 1,
		PixelOffset: pixelOffset,
	}, nil
}
