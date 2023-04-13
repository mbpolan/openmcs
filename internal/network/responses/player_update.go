package responses

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/util"
)

const PlayerUpdateResponseHeader byte = 0x51

const (
	localMoveNoUpdate  byte = 0xFF
	localMoveUnchanged      = 0x00
	localMoveWalk           = 0x01
	localMoveRun            = 0x02
	localMovePosition       = 0x03
)

const (
	updatePlayerWalking     uint16 = 0x400
	updateGraphics                 = 0x100
	updateAnimations               = 0x008
	updateOverheadText             = 0x004
	updateChatMessageText          = 0x080
	updatePlayerInteraction        = 0x001
	updateAppearance               = 0x010
	updateOrientation              = 0x002
	updateDamageSplat              = 0x020
	updateDamageSplatAlt           = 0x200
)

// entityAnimationIDs is the order in which entity appearance animations are written.
var entityAnimationIDs = []model.AnimationID{
	model.AnimationStand,
	model.AnimationStandTurn,
	model.AnimationWalk,
	model.AnimationTurnAbout,
	model.AnimationTurnRight,
	model.AnimationTurnLeft,
	model.AnimationRun,
}

type playerUpdate struct {
	id         int
	mask       uint16
	appearance *entityAppearance
}

type entityAppearance struct {
	name       string
	appearance *model.EntityAppearance
}

// PlayerUpdateResponse contains a game state update.
type PlayerUpdateResponse struct {
	localMoveType       byte
	localPosition       model.Vector3D
	localClearWaypoints bool
	localNeedsUpdate    bool
	updates             []*playerUpdate
}

// NewPlayerUpdateResponse creates a new game state update response.
func NewPlayerUpdateResponse() *PlayerUpdateResponse {
	return &PlayerUpdateResponse{
		localMoveType: localMoveNoUpdate,
	}
}

// SetLocalPlayerNoMovement reports that the local player's state has not changed.
func (p *PlayerUpdateResponse) SetLocalPlayerNoMovement() {
	p.localMoveType = localMoveUnchanged
}

// SetLocalPlayerPosition reports the local player's position in region local coordinates and update status. The
// clearWaypoints flag indicates if the player's current path should be cancelled, such as in the case of the player
// being teleported to a location.
func (p *PlayerUpdateResponse) SetLocalPlayerPosition(pos model.Vector3D, clearWaypoints bool, needsUpdate bool) {
	p.localMoveType = localMovePosition
	p.localPosition = pos
	p.localClearWaypoints = clearWaypoints
	p.localNeedsUpdate = needsUpdate
}

// AddAppearanceUpdate adds a player or NPC appearance update to send to the client.
func (p *PlayerUpdateResponse) AddAppearanceUpdate(playerID int, name string, a *model.EntityAppearance) {
	// is there an existing player update already pending?
	var update *playerUpdate
	for _, u := range p.updates {
		if u.id == playerID {
			update = u
			break
		}
	}

	// first update for this player
	if update == nil {
		update = &playerUpdate{
			id:   playerID,
			mask: 0x00,
		}

		// TODO: maintain ordering consistent with player list indexes
		p.updates = append([]*playerUpdate{update}, p.updates...)
	}

	update.mask |= updateAppearance
	update.appearance = &entityAppearance{
		name:       name,
		appearance: a,
	}
}

// Write writes the contents of the message to a stream.
func (p *PlayerUpdateResponse) Write(w *network.ProtocolWriter) error {
	// since the payload can vary in length, we need to use a buffered write to later compute the size
	bw := network.NewBufferedWriter()
	err := p.writePayload(bw)
	if err != nil {
		return err
	}

	// write packet header
	err = w.WriteByte(PlayerUpdateResponseHeader)
	if err != nil {
		return err
	}

	// now that the payload has been written to a buffer, we can determine its length and write that as 2 bytes
	// note that the packet header is not included in the size itself
	buf, err := bw.Buffer()
	err = w.WriteUint16(uint16(buf.Len()))
	if err != nil {
		return err
	}

	// finally write the payload itself
	_, err = w.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (p *PlayerUpdateResponse) writePayload(w *network.ProtocolWriter) error {
	// prepare a bitset for writing bit-level data
	bs := network.NewBitSet()

	// write local player movement details
	p.writeLocalPlayer(bs)

	// TODO: write 8 bits for the number of other players to update
	bs.SetBits(0, 8)

	// TODO: updates for player list

	// add local player as the last one in the update list
	if p.localMoveType != localMoveNoUpdate {
		bs.SetBits(0x7FF, 11)
	}

	// write bits section first representing local, other and player list updates
	err := bs.Write(w)
	if err != nil {
		return err
	}

	// write each player update block itself
	err = p.writePlayerUpdates(w)
	if err != nil {
		return err
	}

	return nil
}

func (p *PlayerUpdateResponse) writeLocalPlayer(bs *network.BitSet) {
	// first bit is a flag if there is an update for the local player
	if p.localMoveType == localMoveNoUpdate {
		// clear the first bit and bail out since there is nothing else to do
		bs.Clear()
		return
	}

	// set the first bit to indicate we have an update
	bs.Set()

	// two bits represent the local player update type
	bs.SetBits(uint32(p.localMoveType), 2)

	switch p.localMoveType {
	case localMoveUnchanged:
		// nothing to do

	case localMoveWalk:
		// TODO
		panic("not implemented")

	case localMoveRun:
		// TODO
		panic("not implemented")

	case localMovePosition:
		// write 2 bits for the z coordinate
		bs.SetBits(uint32(p.localPosition.Z), 2)

		// write 1 bit each for the clear waypoints and update needed flags
		bs.SetOrClear(p.localClearWaypoints)
		bs.SetOrClear(p.localNeedsUpdate)

		// write 7 bits each for the y and x coordinates
		bs.SetBits(uint32(p.localPosition.Y), 7)
		bs.SetBits(uint32(p.localPosition.X), 7)
	}
}

func (p *PlayerUpdateResponse) writePlayerUpdates(w *network.ProtocolWriter) error {
	for _, u := range p.updates {
		// if the mask cannot fit into a single byte, split it into two
		if u.mask > 0xFF {
			err := w.WriteUint16(u.mask)
			if err != nil {
				return err
			}
		} else {
			err := w.WriteByte(byte(u.mask))
			if err != nil {
				return err
			}
		}

		// write appearance update
		if u.mask&updateAppearance != 0 {
			err := p.writeAppearance(u.appearance, w)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *PlayerUpdateResponse) writeAppearance(ea *entityAppearance, w *network.ProtocolWriter) error {
	a := ea.appearance

	// use a buffered writer since we need to keep track of length of the appearance data
	bw := network.NewBufferedWriter()

	gender := byte(0x00)
	if a.Gender == model.EntityFemale {
		gender = 0x01
	}

	// write 1 byte for the gender
	err := bw.WriteByte(gender)
	if err != nil {
		return err
	}

	// write 1 byte for overhead icon
	err = bw.WriteByte(byte(a.OverheadIconID))
	if err != nil {
		return err
	}

	// write each equipment slot
	for _, id := range a.Equipment {
		// if nothing is equipped at this slot, write one byte only
		if id == 0 {
			err = bw.WriteByte(0)
			if err != nil {
				return err
			}

			continue
		}

		// write 2 bytes for an equipped item
		err = bw.WriteUint16(uint16(id))
		if err != nil {
			return err
		}

		// special case for the first slot: if the appearance is that of an npc, write another 2 bytes to indicate
		// as such and exit the loop
		if a.IsNPCAppearance() {
			err = bw.WriteUint16(uint16(a.NPCAppearance))
			if err != nil {
				return err
			}

			break
		}
	}

	// write each body part color
	for _, color := range a.Body {
		err = bw.WriteByte(byte(color))
		if err != nil {
			return err
		}
	}

	// write each animation id
	for _, i := range entityAnimationIDs {
		animID := a.Animations[i]
		if animID < 0 {
			animID = 0xFFFF
		}

		err = bw.WriteUint16(uint16(animID))
		if err != nil {
			return err
		}
	}

	// convert the name to a long integer
	name := uint64(0)
	validSetLen := uint64(len(util.ValidNameChars))
	for _, ch := range ea.name {
		name *= validSetLen

		if ch >= 'A' && ch <= 'Z' {
			name += uint64((ch + 1) - 'A')
		} else if ch >= 'a' && ch <= 'z' {
			name += uint64((ch + 1) - 'a')
		} else if ch >= '0' && ch <= '9' {
			name += uint64((ch + 27) - '0')
		}
	}

	// normalize the value
	for name%validSetLen == 0 && name != 0 {
		name /= validSetLen
	}

	err = bw.WriteUint64(name)
	if err != nil {
		return err
	}

	// write a byte the combat level
	err = bw.WriteByte(byte(a.CombatLevel))
	if err != nil {
		return err
	}

	// write 2 bytes for the skill level
	err = bw.WriteUint16(uint16(a.SkillLevel))
	if err != nil {
		return err
	}

	buffer, err := bw.Buffer()
	if err != nil {
		return err
	}

	// write a byte for the appearance buffer size. the client expects this to be a negative value.
	err = w.WriteByte(byte(buffer.Len() * -1))
	if err != nil {
		return err
	}

	// write the buffer itself
	_, err = w.Write(buffer.Bytes())
	if err != nil {
		return err
	}

	return nil
}
