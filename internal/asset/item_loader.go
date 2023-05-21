package asset

import (
	"github.com/mbpolan/openmcs/internal/model"
	"io"
)

const (
	opItemEndDefinition         byte = 0x00
	opItemModelID                    = 0x01
	opItemName                       = 0x02
	opItemDescription                = 0x03
	opItemModelZoom                  = 0x04
	opItemModelRotationX             = 0x05
	opItemModelRotationY             = 0x06
	opItemModelOffset1               = 0x07
	opItemModelOffset2               = 0x08
	opItemUnknown1                   = 0x0A
	opItemStackableFlag              = 0x0B
	opItemUnknown2                   = 0x0C
	opItemMembersOnlyFlag            = 0x10
	opItemMaleEquipPrimary           = 0x17
	opItemMaleEquipSecondary         = 0x18
	opItemFemaleEquipPrimary         = 0x19
	opItemFemaleEquipSecondary       = 0x1A
	opItemGroundActionListStart      = 0x1E
	opItemGroundActionListEnd        = 0x22
	opItemActionListStart            = 0x23
	opItemActionListEnd              = 0x27
	opItemColors                     = 0x28
	opItemMaleEquipEmblem            = 0x4E
	opItemFemaleEquipEmblem          = 0x4F
	opItemMaleDialogueID             = 0x5A
	opItemFemaleDialogueID           = 0x5B
	opItemMaleDialogueHatID          = 0x5C
	opItemFemaleDialogueHatID        = 0x5D
	opItemModelRotationZ             = 0x5F
	opItemNoteID                     = 0x61
	opItemNoteTemplateID             = 0x62
	opItemStackableIDListStart       = 0x64
	opItemStackableIDListEnd         = 0x6D
	opItemScaleX                     = 0x6E
	opItemScaleY                     = 0x6F
	opItemScaleZ                     = 0x70
	opItemLightModifier              = 0x71
	opItemShadowModifier             = 0x72
	opItemTeamID                     = 0x73
)

// ItemLoader loads item data from game asset files.
type ItemLoader struct {
	archive *Archive
}

func NewItemLoader(archive *Archive) *ItemLoader {
	return &ItemLoader{
		archive: archive,
	}
}

func (l *ItemLoader) Load() ([]*model.Item, error) {
	// extract the files containing item data
	dataFile, err := l.archive.File("obj.dat")
	if err != nil {
		return nil, err
	}

	idxFile, err := l.archive.File("obj.idx")
	if err != nil {
		return nil, err
	}

	dataReader := NewDataReader(dataFile)
	idxReader := NewDataReader(idxFile)

	// read the number of items in the file
	numItems, err := idxReader.Uint16()
	if err != nil {
		return nil, err
	}

	items := make([]*model.Item, numItems)

	offset := 2
	for i := 0; i < int(numItems); i++ {
		_, err := dataReader.Seek(int64(offset), io.SeekStart)
		if err != nil {
			return nil, err
		}

		item, err := l.readItem(i, dataReader)
		if err != nil {
			return nil, err
		}

		items[i] = item

		// move to the next offset
		nextOffset, err := idxReader.Uint16()
		if err != nil {
			return nil, err
		}

		offset += int(nextOffset)
	}

	return items, nil
}

func (l *ItemLoader) readItem(id int, r *DataReader) (*model.Item, error) {
	item := &model.Item{
		ID: id,
	}

	hasMoreAttributes := true
	for hasMoreAttributes {
		// read the next op code to determine what attribute follows
		op, err := r.ReadByte()
		if err != nil {
			return nil, err
		}

		switch op {
		case opItemEndDefinition:
			// finished reading item definition
			hasMoreAttributes = false

		case opItemModelID:
			// skip a byte containing the item model id
			_, err := r.Uint16()
			if err != nil {
				return nil, err
			}

		case opItemName:
			// read the item name
			item.Name, err = r.String()
			if err != nil {
				return nil, err
			}

		case opItemDescription:
			// read the item description
			item.Description, err = r.String()
			if err != nil {
				return nil, err
			}

		case opItemModelZoom:
			// skip 2 bytes containing zoom
			_, err := r.Uint16()
			if err != nil {
				return nil, err
			}

		case opItemModelRotationX:
			// read 2 bytes containing rotation along x-axis
			v, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			item.Rotation.X = int(v)

		case opItemModelRotationY:
			// read 2 bytes containing rotation along y-axis
			v, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			item.Rotation.Y = int(v)

		case opItemModelOffset1:
			// skip 2 bytes for model offset 1
			_, err = r.Uint16()
			if err != nil {
				return nil, err
			}

		case opItemModelOffset2:
			// skip 2 bytes for model offset 2
			_, err = r.Uint16()
			if err != nil {
				return nil, err
			}

		case opItemUnknown1:
			// skip 2 bytes for an unknown attribute
			_, err = r.Uint16()
			if err != nil {
				return nil, err
			}

		case opItemStackableFlag:
			// flag item as stackable
			item.Stackable = true

		case opItemUnknown2:
			// skip 4 bytes for an unknown attribute
			_, err = r.Uint32()
			if err != nil {
				return nil, err
			}

		case opItemMembersOnlyFlag:
			// flag the item as members-only
			item.MembersOnly = true

		case opItemMaleEquipPrimary:
			// skip 3 bytes for model id (male)
			_, err = r.Uint24()
			if err != nil {
				return nil, err
			}

		case opItemMaleEquipSecondary:
			// skip 2 bytes for secondary model id (male)
			_, err = r.Uint16()
			if err != nil {
				return nil, err
			}

		case opItemFemaleEquipPrimary:
			// skip 3 bytes for model id (female)
			_, err = r.Uint24()
			if err != nil {
				return nil, err
			}

		case opItemFemaleEquipSecondary:
			// skip 2 bytes for secondary model id (female)
			_, err = r.Uint16()
			if err != nil {
				return nil, err
			}

		case opItemColors:
			// read a byte containing color count
			numColors, err := r.Byte()
			if err != nil {
				return nil, err
			}

			// skip 4 bytes per color
			_, err = r.Seek(int64(numColors)*4, io.SeekCurrent)
			if err != nil {
				return nil, err
			}

		case opItemMaleEquipEmblem:
			// skip 2 bytes for equip emblem (male)
			_, err = r.Uint16()
			if err != nil {
				return nil, err
			}

		case opItemFemaleEquipEmblem:
			// skip 2 bytes for equip emblem (female)
			_, err = r.Uint16()
			if err != nil {
				return nil, err
			}

		case opItemMaleDialogueID:
			// skip 2 bytes for dialogue model id (male)
			_, err = r.Uint16()
			if err != nil {
				return nil, err
			}

		case opItemFemaleDialogueID:
			// skip 2 bytes for dialogue model id (female)
			_, err = r.Uint16()
			if err != nil {
				return nil, err
			}

		case opItemMaleDialogueHatID:
			// skip 2 bytes for dialogue hat model id (male)
			_, err = r.Uint16()
			if err != nil {
				return nil, err
			}

		case opItemFemaleDialogueHatID:
			// skip 2 bytes for dialogue hat model id (female)
			_, err = r.Uint16()
			if err != nil {
				return nil, err
			}

		case opItemModelRotationZ:
			// read 2 bytes containing model rotation along z-axis
			v, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			item.Rotation.Z = int(v)

		case opItemNoteID:
			// read 2 bytes for item note model id
			noteID, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			item.NoteID = int(noteID)

		case opItemNoteTemplateID:
			// read 2 bytes for item note template id
			templateID, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			item.NoteTemplateID = int(templateID)

		case opItemScaleX:
			// read 2 bytes for item scale along x-axis
			v, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			item.Scale.X = int(v)

		case opItemScaleY:
			// read 2 bytes for item scale along y-axis
			v, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			item.Scale.Y = int(v)

		case opItemScaleZ:
			// read 2 bytes for item scale along z-axis
			v, err := r.Uint16()
			if err != nil {
				return nil, err
			}

			item.Scale.Z = int(v)

		case opItemLightModifier:
			// skip 1 byte for light modifier
			_, err = r.Byte()
			if err != nil {
				return nil, err
			}

		case opItemShadowModifier:
			// skip 1 byte for shadow modifier
			_, err = r.Byte()
			if err != nil {
				return nil, err
			}

		case opItemTeamID:
			// read 1 byte for team id
			teamID, err := r.Byte()
			if err != nil {
				return nil, err
			}

			item.TeamID = int(teamID)

		default:
			if op >= opItemGroundActionListStart && op <= opItemGroundActionListEnd {
				if len(item.GroundActions) == 0 {
					item.GroundActions = make([]string, 5)
				}

				// read the action name
				action, err := r.String()
				if err != nil {
					return nil, err
				}

				// skip actions that are hidden
				if action != "hidden" {
					idx := int(op - opItemGroundActionListStart)
					item.GroundActions[idx] = action
				}
			} else if op >= opItemActionListStart && op <= opItemActionListEnd {
				if len(item.Actions) == 0 {
					item.Actions = make([]string, 5)
				}

				// read the action name
				action, err := r.String()
				if err != nil {
					return nil, err
				}

				// skip actions that are hidden
				if action != "hidden" {
					idx := int(op - opItemActionListStart)
					item.Actions[idx] = action
				}
			} else if op >= opItemStackableIDListStart && op <= opItemStackableIDListEnd {
				if len(item.Stackables) == 0 {
					item.Stackables = make([]model.ItemStackable, 10)
				}

				// read the stackable id
				stackableID, err := r.Uint16()
				if err != nil {
					return nil, err
				}

				// read the stackable amount
				amount, err := r.Uint16()
				if err != nil {
					return nil, err
				}

				idx := int(op - opItemStackableIDListStart)
				item.Stackables[idx] = model.ItemStackable{
					ID:     int(stackableID),
					Amount: int(amount),
				}
			}
		}
	}

	return item, nil
}
