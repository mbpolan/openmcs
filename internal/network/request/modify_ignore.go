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

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *ModifyIgnoreRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	_, err := r.Uint8()
	if err != nil {
		return err
	}

	// read 8 bytes containing the encoded ignored player name
	name, err := r.Uint64()
	if err != nil {
		return err
	}

	// decode the name
	username, err := util.DecodeName(name)
	if err != nil {
		return err
	}

	p.Username = username
	return nil
}
