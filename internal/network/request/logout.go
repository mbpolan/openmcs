package request

import "github.com/mbpolan/openmcs/internal/network"

const LogoutRequestHeader byte = 0xB9

// LogoutRequest is sent by the client when the player intentionally requests to log out.
type LogoutRequest struct {
	Action int
}

func ReadLogoutRequest(r *network.ProtocolReader) (*LogoutRequest, error) {
	// read 2 bytes containing the reason for the logout
	action, err := r.Uint16()
	if err != nil {
		return nil, err
	}

	return &LogoutRequest{
		Action: int(action),
	}, nil
}