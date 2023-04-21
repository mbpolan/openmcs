package request

import (
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/util"
)

const AddIgnoreRequestHeader byte = 0x85
const RemoveIgnoreRequestHeader byte = 0x4A

// ModifyIgnoreRequest is sent by the client when a player adds or removes another player on their ignore list.
type ModifyIgnoreRequest struct {
	Username string
}

func ReadModifyIgnoreRequest(r *network.ProtocolReader) (*ModifyIgnoreRequest, error) {
	// read 8 bytes containing the encoded ignored player name
	name, err := r.Uint64()
	if err != nil {
		return nil, err
	}

	// decode the name
	username, err := util.DecodeName(name)
	if err != nil {
		return nil, err
	}

	return &ModifyIgnoreRequest{
		Username: username,
	}, nil
}
