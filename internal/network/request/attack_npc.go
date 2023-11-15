package request

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/network"
)

const AttackNPCRequestHeader byte = 0x48

// AttackNPCRequest is sent when the player attacks an NPC.
type AttackNPCRequest struct {
	TargetID int
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *AttackNPCRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	header, err := r.Uint8()
	if err != nil {
		return err
	}

	if header != AttackNPCRequestHeader {
		return fmt.Errorf("invalid header: %2x", header)
	}

	// read 2 bytes for the target npc id
	targetID, err := r.Uint16Alt()
	if err != nil {
		return err
	}

	p.TargetID = int(targetID)
	return nil
}
