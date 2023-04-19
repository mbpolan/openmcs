package response

import (
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/util"
)

// offlineWorldID indicates a player is offline.
const offlineWorldID int = 0x00

const FriendStatusResponseHeader byte = 0x32

// FriendStatusResponse is sent by the server when a friends list player status has changed.
type FriendStatusResponse struct {
	username string
	worldID  int
}

// NewFriendStatusResponse creates a friend status update for an online player.
func NewFriendStatusResponse(username string, worldID int) *FriendStatusResponse {
	return &FriendStatusResponse{
		username: username,
		worldID:  worldID,
	}
}

// NewOfflineFriendStatusResponse creates a friend status update for an offline player.
func NewOfflineFriendStatusResponse(username string) *FriendStatusResponse {
	return &FriendStatusResponse{
		username: username,
		worldID:  offlineWorldID,
	}
}

// Write writes the contents of the message to a stream.
func (p *FriendStatusResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(FriendStatusResponseHeader)
	if err != nil {
		return err
	}

	// encode the player's username and write it as 8 bytes
	name := util.EncodeName(p.username)
	err = w.WriteUint64(name)
	if err != nil {
		return err
	}

	// write 1 byte for the world id
	err = w.WriteUint8(byte(p.worldID))
	if err != nil {
		return err
	}

	return nil
}