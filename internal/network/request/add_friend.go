package request

import (
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/util"
)

const AddFriendRequestHeader byte = 0xBC

// AddFriendRequest is sent by the client when the player adds another player to their friends list.
type AddFriendRequest struct {
	Username string
}

func ReadAddFriendRequest(r *network.ProtocolReader) (*AddFriendRequest, error) {
	// read 8 bytes containing the encoded friend name
	name, err := r.Uint64()
	if err != nil {
		return nil, err
	}

	// decode the name
	username, err := util.DecodeName(name)
	if err != nil {
		return nil, err
	}

	return &AddFriendRequest{
		Username: username,
	}, nil
}
