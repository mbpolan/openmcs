package request

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/network"
)

const InteractWithNPCAction1RequestHeader byte = 0x9B
const InteractWithNPCAction2RequestHeader byte = 0x11
const InteractWithNPCAction3RequestHeader byte = 0x15
const InteractWithNPCAction4RequestHeader byte = 0x12

// InteractWithNPCRequest is sent when the player interacts with an NPC.
type InteractWithNPCRequest struct {
	ActionIndex int
	TargetID    int
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *InteractWithNPCRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	header, err := r.Uint8()
	if err != nil {
		return err
	}

	// read 2 bytes for the target npc id
	targetID, err := r.Uint16Alt()
	if err != nil {
		return err
	}

	// translate the header into an action index
	switch header {
	case InteractWithNPCAction1RequestHeader:
		p.ActionIndex = 0
	case InteractWithNPCAction2RequestHeader:
		p.ActionIndex = 1
	case InteractWithNPCAction3RequestHeader:
		p.ActionIndex = 2
	case InteractWithNPCAction4RequestHeader:
		p.ActionIndex = 3
	default:
		return fmt.Errorf("unexpected interact with NPC header: %2x", header)
	}

	p.TargetID = int(targetID)
	return nil
}
