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

	// translate the header into an action index and read 2 bytes for the target npc id. the format of the id
	// varies depending on the packet
	var targetID uint16
	switch header {
	case InteractWithNPCAction1RequestHeader:
		p.ActionIndex = 0

		targetID, err = r.Uint16LE()
		if err != nil {
			return err
		}
	case InteractWithNPCAction2RequestHeader:
		p.ActionIndex = 1

		targetID, err = r.Uint16LEAlt()
		if err != nil {
			return err
		}
	case InteractWithNPCAction3RequestHeader:
		p.ActionIndex = 2

		targetID, err = r.Uint16()
		if err != nil {
			return err
		}
	case InteractWithNPCAction4RequestHeader:
		p.ActionIndex = 3

		targetID, err = r.Uint16LE()
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unexpected interact with NPC header: %2x", header)
	}

	p.TargetID = int(targetID)
	return nil
}
