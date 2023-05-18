package request

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
)

const CharacterDesignRequestHeader byte = 0x65

// CharacterDesignRequest is sent by the client when a player submits their character design.
type CharacterDesignRequest struct {
	Gender     model.EntityGender
	BodyColors []int
	Base       model.EntityBase
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *CharacterDesignRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	b, err := r.Uint8()
	if err != nil {
		return err
	}

	if b != CharacterDesignRequestHeader {
		return fmt.Errorf("unexpected packet header: %x", b)
	}

	// read 1 byte for the gender
	b, err = r.Uint8()
	if err != nil {
		return err
	}

	if b == 0x00 {
		p.Gender = model.EntityMale
	} else if b == 0x01 {
		p.Gender = model.EntityFemale
	} else {
		return fmt.Errorf("invalid character gender byte: %x", b)
	}

	// read 7 bytes for the entity base model
	for i := 0; i < 7; i++ {
		b, err = r.Uint8()
		if err != nil {
			return err
		}

		// if a slot is unused, 0xFF will be sent instead of the appearance id
		id := int(b)
		if id == 0xFF {
			id = -1
		}

		switch i {
		case 0:
			p.Base.Head = id
		case 1:
			p.Base.Face = id
		case 2:
			p.Base.Body = id
		case 3:
			p.Base.Arms = id
		case 4:
			p.Base.Hands = id
		case 5:
			p.Base.Legs = id
		case 6:
			p.Base.Feet = id
		}
	}

	// read 5 bytes for the body colors
	p.BodyColors = make([]int, 5)
	for i := 0; i < 5; i++ {
		b, err = r.Uint8()
		if err != nil {
			return err
		}

		p.BodyColors[i] = int(b)
	}

	return nil
}
