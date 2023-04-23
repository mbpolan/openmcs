package request

import "github.com/mbpolan/openmcs/internal/network"

const InterfaceActionRequestHeader byte = 0xB9

// InterfaceActionRequest is sent by the client when the player interacts with an interface.
type InterfaceActionRequest struct {
	Action int
}

func ReadInterfaceActionRequest(r *network.ProtocolReader) (*InterfaceActionRequest, error) {
	// read 2 bytes containing the interface action id
	action, err := r.Uint16()
	if err != nil {
		return nil, err
	}

	return &InterfaceActionRequest{
		Action: int(action),
	}, nil
}
