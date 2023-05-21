package asset

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/pkg/errors"
)

// InterfaceLoader loads interface data from game asset files.
type InterfaceLoader struct {
	archive *Archive
}

// NewInterfaceLoader returns a new loader for interface data.
func NewInterfaceLoader(archive *Archive) *InterfaceLoader {
	return &InterfaceLoader{
		archive: archive,
	}
}

// Load returns a slice of interface data extracted from game assets, or an error if the data could not be read.
func (l *InterfaceLoader) Load() ([]*model.Interface, error) {
	dataFile, err := l.archive.File("data")
	if err != nil {
		return nil, err
	}

	reader := NewDataReader(dataFile)

	// read the maximum of the interface ids
	maxInterfaceID, err := reader.Uint16()
	if err != nil {
		return nil, err
	}

	interfaces := map[int]*model.Interface{}
	parentIDs := map[int]int{}
	childIDs := map[int][]int{}

	// read data for each interface
	lastParentID := -1
	for reader.HasMore() {
		var id int

		// read 2 bytes for the interface id
		v, err := reader.Uint16()
		if err != nil {
			return nil, err
		}

		// if the id is 0xFFFF, then this interface marks the beginning of a new parent
		if v == 0xFFFF {
			v, err = reader.Uint16()
			if err != nil {
				return nil, err
			}

			lastParentID = int(v)

			v, err = reader.Uint16()
			if err != nil {
				return nil, err
			}

			id = int(v)
		} else {
			id = int(v)
		}

		// record the parent for this interface
		parentIDs[id] = lastParentID

		// bounds check both ids
		if id < 0 || id > int(maxInterfaceID)-1 {
			return nil, fmt.Errorf("interface id out of bounds: %d", id)
		}
		if lastParentID < 0 || lastParentID > int(maxInterfaceID)-1 {
			return nil, fmt.Errorf("parent interface id out of bounds: %d", lastParentID)
		}

		// read the interface data
		inf, children, err := l.readInterface(id, lastParentID, reader)
		if err != nil {
			return nil, errors.Wrapf(err, "failed loading interface %d", id)
		}

		interfaces[id] = inf
		childIDs[id] = children
	}

	// resolve parent and child interfaces
	for _, inf := range interfaces {
		// check if this interface has a parent interface
		parentID, ok := parentIDs[inf.ID]
		if !ok || parentID == -1 || parentID == inf.ID {
			continue
		}

		inf.Parent, ok = interfaces[parentID]
		if !ok {
			return nil, fmt.Errorf("missing parent %d for interface %d", parentID, inf.ID)
		}

		children, ok := childIDs[inf.ID]
		if !ok {
			continue
		}

		for _, childID := range children {
			child, ok := interfaces[childID]
			if !ok {
				return nil, fmt.Errorf("missing child %d for interface %d", parentID, inf.ID)
			}

			inf.Children = append(inf.Children, child)
		}
	}

	var infs []*model.Interface
	for _, inf := range interfaces {
		infs = append(infs, inf)
	}

	return infs, nil
}

func (l *InterfaceLoader) readInterface(id, parentID int, reader *DataReader) (*model.Interface, []int, error) {
	inf := &model.Interface{
		ID:      id,
		OpCodes: map[int][]int{},
	}

	var childIDs []int

	// read 1 byte for the interface type
	infType, err := reader.Byte()
	if err != nil {
		return nil, nil, err
	}

	// read 1 byte for the action type
	actionType, err := reader.Byte()
	if err != nil {
		return nil, nil, err
	}

	// read 2 bytes for the content type
	_, err = reader.Uint16()
	if err != nil {
		return nil, nil, err
	}

	// skip 2 bytes for the width
	_, err = reader.Uint16()
	if err != nil {
		return nil, nil, err
	}

	// skip 2 bytes for the height
	_, err = reader.Uint16()
	if err != nil {
		return nil, nil, err
	}

	// skip 1 byte for the alpha color
	_, err = reader.Byte()
	if err != nil {
		return nil, nil, err
	}

	// read 1 byte if this interface has a popup
	b, err := reader.Byte()
	if err != nil {
		return nil, nil, err
	}

	// skip another byte if there is a popup
	if b != 0x00 {
		_, err := reader.Byte()
		if err != nil {
			return nil, nil, err
		}
	}

	// read 1 byte for the number of interface conditions
	b, err = reader.Byte()
	if err != nil {
		return nil, nil, err
	}

	// read data for each interface condition
	inf.Conditions = make([]model.InterfaceCondition, int(b))
	for i := 0; i < int(b); i++ {
		// read 1 byte for the condition type
		t, err := reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		// read 2 bytes for the condition value
		v, err := reader.Uint16()
		if err != nil {
			return nil, nil, err
		}

		inf.Conditions[i] = model.InterfaceCondition{
			Type:  int(t),
			Value: int(v),
		}
	}

	// read 1 byte for the number of op codes
	b, err = reader.Byte()
	if err != nil {
		return nil, nil, err
	}

	// read each op code and its sub codes
	for i := 0; i < int(b); i++ {
		// read 2 bytes for the number of sub codes
		n, err := reader.Uint16()
		if err != nil {
			return nil, nil, err
		}

		codes := make([]int, int(n))
		// read each sub code for this op code
		for j := 0; j < int(n); j++ {
			// read 2 bytes for the sub code
			code, err := reader.Uint16()
			if err != nil {
				return nil, nil, err
			}

			codes[j] = int(code)
		}

		inf.OpCodes[i] = codes
	}

	// read additional data depending on the type of interface
	switch infType {
	case 0:
		// parent interface containing one or more child interfaces

		// skip 2 bytes for the maximum scroll size
		_, err = reader.Uint16()
		if err != nil {
			return nil, nil, err
		}

		// skip 1 byte for a flag if this interface appears only on mouse hover
		_, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		// read 2 bytes for the number of child interfaces
		n, err := reader.Uint16()
		if err != nil {
			return nil, nil, err
		}

		// read each child interface pointer
		childIDs = make([]int, int(n))
		for i := 0; i < int(n); i++ {
			// read 2 bytes for the child interface id
			ch, err := reader.Uint16()
			if err != nil {
				return nil, nil, err
			}

			// skip 2 bytes for the x-coordinate offset
			_, err = reader.Uint16()
			if err != nil {
				return nil, nil, err
			}

			// skip 2 bytes for the y-coordinate offset
			_, err = reader.Uint16()
			if err != nil {
				return nil, nil, err
			}

			childIDs[i] = int(ch)
		}

		// read selected action and spell if supported
		if actionType == 0x02 {
			err = l.readSpellActionSection(reader)
			if err != nil {
				return nil, nil, err
			}
		}

	case 1:
		// unknown/unused interface type

		// skip 3 bytes
		_, err := reader.Uint24()
		if err != nil {
			return nil, nil, err
		}

		// skip 1 byte for a flag if text should be centered
		_, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		// skip 1 byte for the font id
		_, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		// skip 1 byte for a flag if text should be drawn with a shadow
		_, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		// skip 4 bytes for the default color
		_, err = reader.Uint32()
		if err != nil {
			return nil, nil, err
		}

		// read selected action and spell if supported
		if actionType == 0x02 {
			err = l.readSpellActionSection(reader)
			if err != nil {
				return nil, nil, err
			}
		}

	case 2:
		// interface for displaying/managing inventories of items

		// skip 1 byte if items can be moved
		_, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		// skip 1 byte for a flag if this interface supports inventory items
		_, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		// skip 1 byte for a flag if items can be used in this interface
		_, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		// skip 1 byte for a flag if items can be drag-n-dropped in this interface
		_, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		// skip 1 byte for sprite column padding
		_, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		// skip 1 byte for sprite row padding
		_, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		// read data for each sprite slot
		for i := 0; i < 20; i++ {
			// read 1 byte to determine if a sprite is used in this slot
			used, err := reader.Byte()
			if err != nil {
				return nil, nil, err
			}

			// no more data for this slot if it has no sprite
			if used != 0x01 {
				continue
			}

			// skip 2 bytes for the x-coordinate
			_, err = reader.Uint16()
			if err != nil {
				return nil, nil, err
			}

			// skip 2 bytes for the y-coordinate
			_, err = reader.Uint16()
			if err != nil {
				return nil, nil, err
			}

			// skip a string for the sprite name
			_, err = reader.String()
			if err != nil {
				return nil, nil, err
			}
		}

		// read 5 strings for interface actions
		for i := 0; i < 5; i++ {
			action, err := reader.String()
			if err != nil {
				return nil, nil, err
			}

			inf.Actions = append(inf.Actions, action)
		}

		// read selected action, spell and target
		err = l.readSpellActionSection(reader)
		if err != nil {
			return nil, nil, err
		}

	case 3:
		// blank interface

		// skip 1 byte for a flag if this interface has a fill color
		_, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		// skip 4 bytes for the fill color
		_, err = reader.Uint32()
		if err != nil {
			return nil, nil, err
		}

		// skip 4 bytes for the active color
		_, err = reader.Uint32()
		if err != nil {
			return nil, nil, err
		}

		// skip 4 bytes for the default hover color
		_, err = reader.Uint32()
		if err != nil {
			return nil, nil, err
		}

		// skip 4 bytes for the active hovered color
		_, err = reader.Uint32()
		if err != nil {
			return nil, nil, err
		}

		// read selected action and spell if supported
		if actionType == 0x02 {
			err = l.readSpellActionSection(reader)
			if err != nil {
				return nil, nil, err
			}
		}

	case 4:
		// interface that displays text

		// skip 1 byte for a flag if text should be centered
		_, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		// skip 1 byte for the font id
		_, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		// skip 1 byte for a flag if text should be drawn with a shadow
		_, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		// skip 1 string for the default text
		_, err = reader.String()
		if err != nil {
			return nil, nil, err
		}

		// skip 1 string for the active text
		_, err = reader.String()
		if err != nil {
			return nil, nil, err
		}

		// skip 4 bytes for the default color
		_, err = reader.Uint32()
		if err != nil {
			return nil, nil, err
		}

		// skip 4 bytes for the active color
		_, err = reader.Uint32()
		if err != nil {
			return nil, nil, err
		}

		// skip 4 bytes for the default hover color
		_, err = reader.Uint32()
		if err != nil {
			return nil, nil, err
		}

		// skip 4 bytes for the active hovered color
		_, err = reader.Uint32()
		if err != nil {
			return nil, nil, err
		}

		// read selected action and spell if supported
		if actionType == 0x02 {
			err = l.readSpellActionSection(reader)
			if err != nil {
				return nil, nil, err
			}
		}

	case 5:
		// interface for displaying sprites

		// skip 1 string for the default sprite name
		_, err = reader.String()

		// skip 1 string for the active sprite name
		_, err = reader.String()

		// read selected action and spell if supported
		if actionType == 0x02 {
			err = l.readSpellActionSection(reader)
			if err != nil {
				return nil, nil, err
			}
		}

	case 6:
		// interface for displaying a 3d model

		// read 1 byte for the default model id
		i, err := reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		if i != 0x00 {
			// skip 1 byte for the low byte of the default model id
			_, err = reader.Byte()
			if err != nil {
				return nil, nil, err
			}
		}

		// read 1 byte for the active model id
		i, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		if i != 0x00 {
			// skip 1 byte for the low byte of the active model id
			_, err = reader.Byte()
			if err != nil {
				return nil, nil, err
			}
		}

		// read 1 byte for the default animation id
		i, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		if i != 0x00 {
			// skip 1 byte for the low byte of the default animation id
			_, err = reader.Byte()
			if err != nil {
				return nil, nil, err
			}
		}

		// read 1 byte for the active animation id
		i, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		if i != 0x00 {
			// skip 1 byte for the low byte of the active animation id
			_, err = reader.Byte()
			if err != nil {
				return nil, nil, err
			}
		}

		// skip 2 bytes for the model zoom
		_, err = reader.Uint16()
		if err != nil {
			return nil, nil, err
		}

		// skip 2 bytes for the model x-rotation
		_, err = reader.Uint16()
		if err != nil {
			return nil, nil, err
		}

		// skip 2 bytes for the model y-rotation
		_, err = reader.Uint16()
		if err != nil {
			return nil, nil, err
		}

		// read selected action and spell if supported
		if actionType == 0x02 {
			err = l.readSpellActionSection(reader)
			if err != nil {
				return nil, nil, err
			}
		}

	case 7:
		// interface that displays an inventory of items

		// skip 1 byte for a flag if the interface shows centered text
		_, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		// skip 1 byte for the font id
		_, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		// skip 1 byte for a flag if the text should have a shadow
		_, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		// skip 4 bytes for the default color
		_, err = reader.Uint32()
		if err != nil {
			return nil, nil, err
		}

		// skip 2 bytes for sprite column padding
		_, err = reader.Uint16()
		if err != nil {
			return nil, nil, err
		}

		// skip 2 bytes for sprite row padding
		_, err = reader.Uint16()
		if err != nil {
			return nil, nil, err
		}

		// skip 1 byte for a flag if this interface contains inventory items
		_, err = reader.Byte()
		if err != nil {
			return nil, nil, err
		}

		// read 5 strings for supported interface actions
		for i := 0; i < 5; i++ {
			action, err := reader.String()
			if err != nil {
				return nil, nil, err
			}

			inf.Actions = append(inf.Actions, action)
		}

		// read selected action and spell if supported
		if actionType == 0x02 {
			err = l.readSpellActionSection(reader)
			if err != nil {
				return nil, nil, err
			}
		}

	case 8:
		// interface for displaying text

		// skip 1 string for the default text
		_, err = reader.String()
		if err != nil {
			return nil, nil, err
		}
	}

	// read tooltip if supported
	if l.hasTooltip(actionType) {
		err = l.readTooltip(reader)
		if err != nil {
			return nil, nil, err
		}
	}

	return inf, childIDs, nil
}

// readSpellActionSection reads data from the reader for interfaces that support a selected action, spell and target.
func (l *InterfaceLoader) readSpellActionSection(reader *DataReader) error {
	// skip 1 string for the selected action name
	_, err := reader.String()
	if err != nil {
		return err
	}

	// skip 1 string for the spell name
	_, err = reader.String()
	if err != nil {
		return err
	}

	// skip 2 bytes for spell target id
	_, err = reader.Uint16()
	if err != nil {
		return err
	}

	return nil
}

// hasTooltip returns true if an interface with an action type has a tooltip.
func (l *InterfaceLoader) hasTooltip(actionType byte) bool {
	return actionType == 0x01 || actionType == 0x04 || actionType == 0x05 || actionType == 0x06
}

// readToolTip reads data from the reader for interfaces that support a tooltip.
func (l *InterfaceLoader) readTooltip(reader *DataReader) error {
	// skip 1 string for the tooltip
	_, err := reader.String()
	if err != nil {
		return err
	}

	return nil
}
