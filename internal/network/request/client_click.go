package request

import "github.com/mbpolan/openmcs/internal/network"

const ClientClickRequestHeader byte = 0xF1

// ClientClickRequest is sent by the client when the player clicks on an area of the client window.
type ClientClickRequest struct {
	TimeSince   uint32
	RightClick  bool
	PixelOffset uint32
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *ClientClickRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	_, err := r.Uint8()
	if err != nil {
		return err
	}

	// read a single integer that contains packed data
	v, err := r.Uint32()
	if err != nil {
		return err
	}

	// time since last click is contained in the high 12 bits
	timeSince := v & 0xFFF

	// left or right click is in the 13th bit
	rightClick := v & 0x80000

	// pixel offset is in the remaining bits
	pixelOffset := v & 0x7FFFF

	p.TimeSince = timeSince
	p.RightClick = rightClick == 1
	p.PixelOffset = pixelOffset
	return nil
}
