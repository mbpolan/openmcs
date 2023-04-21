package response

import (
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/util"
)

const IgnoredListResponseHeader byte = 0xD6

// IgnoredListResponse is sent by the server to tell the client about all the players on a player's ignored list.
type IgnoredListResponse struct {
	usernames []string
}

// NewIgnoredListResponse creates a response containing a player's entire ignored list.
func NewIgnoredListResponse(usernames []string) *IgnoredListResponse {
	return &IgnoredListResponse{
		usernames: usernames,
	}
}

// Write writes the contents of the message to a stream.
func (p *IgnoredListResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(IgnoredListResponseHeader)
	if err != nil {
		return err
	}

	// write 2 bytes for the packet size
	err = w.WriteUint16(uint16(len(p.usernames) * 8))
	if err != nil {
		return err
	}

	// write each username as an encoded, 8 byte integer
	for _, username := range p.usernames {
		encoded := util.EncodeName(username)
		err := w.WriteUint64(encoded)
		if err != nil {
			return err
		}
	}

	return nil
}
