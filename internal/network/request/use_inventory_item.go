package request

import "github.com/mbpolan/openmcs/internal/network"

const UseInventoryItemsRequestHeader byte = 0x35

// UseInventoryItemsRequest is sent by the client when the player uses an inventory item on another item.
type UseInventoryItemsRequest struct {
	SourceSlotID      int
	TargetSlotID      int
	SourceItemID      int
	TargetItemID      int
	SourceInterfaceID int
	TargetInterfaceID int
}

func (p *UseInventoryItemsRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	_, err := r.Uint8()
	if err != nil {
		return err
	}

	// read 2 byte for the target item slot id
	targetSlotID, err := r.Uint16()

	// read 2 bytes for the source item slot id
	sourceSlotID, err := r.Uint16Alt()

	// read 2 bytes for the target item id
	targetItemID, err := r.Uint16LEAlt()

	// read 2 bytes for the source item interface id
	sourceInterfaceID, err := r.Uint16()

	// read 2 bytes for the source item id
	sourceItemID, err := r.Uint16LE()

	// read 2 bytes for the target item interface id

}
