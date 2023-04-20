package request

import (
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/util"
)

const AddFriendRequestHeader byte = 0xBC
const RemoveFriendRequestHeader byte = 0xD7

// ModifyFriendRequest is sent by the client when the player adds or removes another player on their friends list.
type ModifyFriendRequest struct {
	Username string
}

func ReadModifyFriendRequest(r *network.ProtocolReader) (*ModifyFriendRequest, error) {
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

	return &ModifyFriendRequest{
		Username: username,
	}, nil
}
