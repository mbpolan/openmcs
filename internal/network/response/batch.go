package response

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
)

const BatchResponseHeader byte = 0x3C

// BatchResponse is sent by the server to communicate multiple messages to the client.
type BatchResponse struct {
	playerRegionRelative model.Vector2D
	responses            []Response
}

// NewBatchResponse creates a response containing multiple batched responses.
func NewBatchResponse(playerRegionRelative model.Vector2D, responses []Response) *BatchResponse {
	return &BatchResponse{
		playerRegionRelative: playerRegionRelative,
		responses:            responses,
	}
}

// Write writes the contents of the message to a stream.
func (p *BatchResponse) Write(w *network.ProtocolWriter) error {
	// use a buffered writer since we need to compute the packet size
	bw := network.NewBufferedWriter()

	// write 1 byte for the player y-coordinate
	err := bw.WriteUint8(byte(p.playerRegionRelative.Y))
	if err != nil {
		return err
	}

	// write 1 byte for the player x-coordinate (inverted)
	err = bw.WriteUint8(byte(p.playerRegionRelative.X * -1))
	if err != nil {
		return err
	}

	// write each response to the buffered writer
	for _, resp := range p.responses {
		err := resp.Write(bw)
		if err != nil {
			return err
		}
	}

	// write the packet to the stream now that its contents are complete

	// write packet header
	err = w.WriteUint8(BatchResponseHeader)
	if err != nil {
		return err
	}

	// write 2 bytes for the length of the buffered data
	buf, err := bw.Buffer()
	err = w.WriteUint16(uint16(buf.Len()))
	if err != nil {
		return err
	}

	// finally write the payload itself
	_, err = w.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}
