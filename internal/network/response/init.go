package response

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/network/common"
)

const initAccepted byte = 0x00
const initLoggedIn byte = 0x02
const initSuccessReset byte = 0x0F

// InitFailureCode enumerates various error conditions that result in a failed initialization.
type InitFailureCode byte

const (
	InitInvalidUsername InitFailureCode = 0x03
	InitAccountDisabled                 = 0x04
	InitAccountLoggedIn                 = 0x05
	InitGameUpdated                     = 0x06
	InitServerFull                      = 0x07
	InitServerOffline                   = 0x08
	InitTooManyAttempts                 = 0x09
	InitInvalidSid                      = 0x0A
	InitLoginRejected                   = 0x0B
	InitNoSubscription                  = 0x0C
	InitLoginFailed                     = 0x0D
	InitMaintenance                     = 0x0E
	InitSlowDown                        = 0x11
	InitInvalidLocation                 = 0x12
	InitWrongServer                     = 0x14
	InitCountDown                       = 0x15
)

// InitResponse is sent by the response in response to a client's initialization request.
type InitResponse struct {
	code          byte
	playerType    byte
	playerFlagged byte
	sessionKey    uint64
}

// NewAcceptedInitResponse creates a response confirming a player's connection was accepted.
func NewAcceptedInitResponse(sessionKey uint64) *InitResponse {
	return &InitResponse{
		code:       initAccepted,
		sessionKey: sessionKey,
	}
}

// NewLoggedInInitResponse creates a response confirming that a player's has been authenticated.
func NewLoggedInInitResponse(playerType model.PlayerType, playerFlagged bool) *InitResponse {
	var flagged byte = 0x00
	if playerFlagged {
		flagged = 0x01
	}

	pType := common.PlayerTypeCode(playerType)

	return &InitResponse{
		code:          initLoggedIn,
		playerType:    pType,
		playerFlagged: flagged,
		sessionKey:    0,
	}
}

// NewFailedInitResponse creates a response indicating the player's login was rejected.
func NewFailedInitResponse(code InitFailureCode) *InitResponse {
	return &InitResponse{
		code: byte(code),
	}
}

// Write writes the contents of the message to a stream.
func (p *InitResponse) Write(w *network.ProtocolWriter) error {
	// write the result code first
	err := w.WriteUint8(p.code)
	if err != nil {
		return err
	}

	// certain response codes contain additional data
	if p.code == initAccepted {
		// write the session key
		err = w.WriteUint64(p.sessionKey)
		if err != nil {
			return err
		}
	} else if p.code == initLoggedIn {
		err = w.WriteUint8(p.playerType)
		if err != nil {
			return err
		}

		err = w.WriteUint8(p.playerFlagged)
		if err != nil {
			return err
		}
	}

	return nil
}
