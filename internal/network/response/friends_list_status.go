package response

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
)

// friendsListStatusCodes maps friends list statuses to protocol identifiers.
var friendsListStatusCodes = map[model.FriendsListStatus]byte{
	model.FriendsListStatusLoading: 0x00,
	model.FriendsListStatusPending: 0x01,
	model.FriendsListStatusLoaded:  0x02,
}

const FriendsListStatusResponseHeader byte = 0xDD

// FriendsListStatusResponse is sent by the server to indicate if a player's friends list has loaded.
type FriendsListStatusResponse struct {
	status byte
}

// NewFriendsListStatusResponse creates a friends list status response.
func NewFriendsListStatusResponse(status model.FriendsListStatus) *FriendsListStatusResponse {
	return &FriendsListStatusResponse{
		status: friendsListStatusCodes[status],
	}
}

// Write writes the contents of the message to a stream.
func (p *FriendsListStatusResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(FriendsListStatusResponseHeader)
	if err != nil {
		return err
	}

	// write 1 byte for the status
	err = w.WriteUint8(p.status)
	if err != nil {
		return err
	}

	return nil
}
