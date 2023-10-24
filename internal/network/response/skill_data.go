package response

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
)

const SkillDataResponseHeader byte = 0x86

// SkillDataResponse is sent by the server to inform the client about a player's skill data for a single skill.
type SkillDataResponse struct {
	id         int
	level      int
	experience int
}

// NewSkillDataResponse creates a new skill data response.
func NewSkillDataResponse(skill *model.Skill) *SkillDataResponse {
	return &SkillDataResponse{
		id:         int(skill.Type),
		level:      skill.StatLevel,
		experience: int(skill.Experience),
	}
}

// Write writes the contents of the message to a stream.
func (p *SkillDataResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(SkillDataResponseHeader)
	if err != nil {
		return err
	}

	// write 1 byte for the skill id
	err = w.WriteUint8(byte(p.id))
	if err != nil {
		return err
	}

	// split the experience points into a sequence of bytes since the client expects them to be ordered in
	// a really weird way
	exp1 := uint8(p.experience >> 8)
	exp2 := uint8(p.experience)
	exp3 := uint8(p.experience >> 24)
	exp4 := uint8(p.experience >> 16)

	// write 4 bytes for the experience
	err = w.WriteUint8(exp1)
	if err != nil {
		return err
	}

	err = w.WriteUint8(exp2)
	if err != nil {
		return err
	}

	err = w.WriteUint8(exp3)
	if err != nil {
		return err
	}

	err = w.WriteUint8(exp4)
	if err != nil {
		return err
	}

	// write 1 byte for the skill level
	err = w.WriteUint8(byte(p.level))
	if err != nil {
		return err
	}

	return nil
}
