package response

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/network/common"
	"github.com/mbpolan/openmcs/internal/util"
	"sort"
)

const PlayerUpdateResponseHeader byte = 0x51

// localPlayerID identifies the local player.
const localPlayerID = 0x7FF

const (
	playerMoveNoUpdate  byte = 0xFF
	playerMoveUnchanged      = 0x00
	playerMoveWalk           = 0x01
	playerMoveRun            = 0x02
	playerMovePosition       = 0x03
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

var directionCodes = map[model.Direction]byte{
	model.DirectionNorth:     0x01,
	model.DirectionNorthWest: 0x00,
	model.DirectionWest:      0x03,
	model.DirectionSouthWest: 0x05,
	model.DirectionSouth:     0x06,
	model.DirectionSouthEast: 0x07,
	model.DirectionEast:      0x04,
	model.DirectionNorthEast: 0x02,
}

type playerUpdate struct {
	mask        uint16
	appearance  *entityAppearance
	chatMessage *model.ChatMessage
}

type entityAppearance struct {
	name       string
	appearance *model.EntityAppearance
}

type trackedPlayer struct {
	observed       bool
	clearWaypoints bool
	pos            model.Vector2D
	update         *playerUpdate
	movement       *playerMovement
}

type playerMovement struct {
	moveType       byte
	position       model.Vector3D
	clearWaypoints bool
	walkDirection  model.Direction
}

// PlayerUpdateResponse contains a game state update.
type PlayerUpdateResponse struct {
	localPlayerID int
	local         *playerMovement
	list          map[int]*trackedPlayer
}

// NewPlayerUpdateResponse creates a new game state update response for a player.
func NewPlayerUpdateResponse(localPlayerID int) *PlayerUpdateResponse {
	return &PlayerUpdateResponse{
		localPlayerID: localPlayerID,
		local: &playerMovement{
			moveType: playerMoveNoUpdate,
		},
		list: map[int]*trackedPlayer{},
	}
}

// Tracking returns true if this response already contains an update for a player.
func (p *PlayerUpdateResponse) Tracking(playerID int) bool {
	if playerID == p.localPlayerID {
		return true
	}

	_, ok := p.list[playerID]
	return ok
}

// SyncLocalMovement adds the other player's movement update to this player's update.
func (p *PlayerUpdateResponse) SyncLocalMovement(playerID int, other *PlayerUpdateResponse) {
	p.ensurePlayer(playerID).movement = other.local
}

// SetLocalPlayerNoMovement reports that the local player's state has not changed.
func (p *PlayerUpdateResponse) SetLocalPlayerNoMovement() {
	p.local.moveType = playerMoveUnchanged

	// this requires an update to be included
	p.list[localPlayerID] = &trackedPlayer{
		update: &playerUpdate{},
	}
}

// SetLocalPlayerWalk reports that the local player is walking in a particular direction.
func (p *PlayerUpdateResponse) SetLocalPlayerWalk(dir model.Direction) {
	p.local.moveType = playerMoveWalk
	p.local.walkDirection = dir

	// start tracking the local player
	if p.list[localPlayerID] == nil {
		p.list[localPlayerID] = &trackedPlayer{}
	}
}

// SetLocalPlayerPosition reports the local player's position in region local coordinates. The clearWaypoints flag
// indicates if the player's current path should be cancelled, such as in the case of the player being teleported to
// a location.
func (p *PlayerUpdateResponse) SetLocalPlayerPosition(pos model.Vector3D, clearWaypoints bool) {
	p.local.moveType = playerMovePosition
	p.local.position = pos
	p.local.clearWaypoints = clearWaypoints
	p.list[localPlayerID] = &trackedPlayer{}
}

func (p *PlayerUpdateResponse) AddOtherPlayerNoUpdate(playerID int) {
	p.ensurePlayer(playerID).movement = &playerMovement{
		moveType: playerMoveNoUpdate,
	}
}

// AddOtherPlayerWalk reports that another player is walking in a particular direction.
func (p *PlayerUpdateResponse) AddOtherPlayerWalk(playerID int, dir model.Direction) {
	p.ensurePlayer(playerID).movement = &playerMovement{
		moveType:       playerMoveWalk,
		position:       model.Vector3D{},
		clearWaypoints: true,
		walkDirection:  dir,
	}
}

// AddToPlayerList adds a tracked player to the local player list. The position should be relative to the local player.
func (p *PlayerUpdateResponse) AddToPlayerList(playerID int, posOffset model.Vector2D, clearWaypoints, observed bool) {
	p.list[playerID] = &trackedPlayer{
		observed:       observed,
		clearWaypoints: clearWaypoints,
		pos:            posOffset,
		// this requires an update to be included
		update: &playerUpdate{},
	}
}

// AddAppearanceUpdate adds a player or NPC appearance update to send to the client.
func (p *PlayerUpdateResponse) AddAppearanceUpdate(playerID int, name string, a *model.EntityAppearance) {
	// if this appearance is for the local player, use the well-known id instead
	id := playerID
	if id == p.localPlayerID {
		id = localPlayerID
	}

	update := p.ensureUpdate(id)
	update.mask |= updateAppearance
	update.appearance = &entityAppearance{
		name:       name,
		appearance: a,
	}
}

// AddChatMessage adds a chat message that was posted by another player.
func (p *PlayerUpdateResponse) AddChatMessage(playerID int, msg *model.ChatMessage) {
	// do not include chat messages for the local player
	if playerID == p.localPlayerID {
		return
	}

	update := p.ensureUpdate(playerID)
	update.mask |= updateChatMessageText
	update.chatMessage = msg
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
	err = w.WriteUint8(PlayerUpdateResponseHeader)
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

func (p *PlayerUpdateResponse) ensurePlayer(playerID int) *trackedPlayer {
	pl, ok := p.list[playerID]
	if !ok {
		pl = &trackedPlayer{}
		p.list[playerID] = pl
	}

	return pl
}

// ensureUpdate returns a pointer to a player's pending updates, or creates an empty update if none were prepared.
func (p *PlayerUpdateResponse) ensureUpdate(playerID int) *playerUpdate {
	pl := p.ensurePlayer(playerID)
	if pl.update == nil {
		pl.update = &playerUpdate{}
	}

	return pl.update
}

func (p *PlayerUpdateResponse) writePayload(w *network.ProtocolWriter) error {
	// collect all players and updates and order them by the player id
	var playerIDs []int
	for k, _ := range p.list {
		playerIDs = append(playerIDs, k)
	}

	sort.Slice(playerIDs, func(i, j int) bool {
		return playerIDs[i] < playerIDs[j]
	})

	// prepare a bitset for writing bit-level data
	bs := network.NewBitSet()

	// write local player movement details
	p.writeLocalPlayer(bs)

	// write 8 bits for the number of other players to update
	p.writeOtherMovements(bs)

	// write the local player list
	p.writePlayerList(playerIDs, bs)

	// write bits section first representing local, other and player list updates
	err := bs.Write(w)
	if err != nil {
		return err
	}

	// write each player update block itself
	err = p.writePlayerUpdates(playerIDs, w)
	if err != nil {
		return err
	}

	return nil
}

func (p *PlayerUpdateResponse) writeLocalPlayer(bs *network.BitSet) {
	moveType := p.local.moveType

	// first bit is a flag if there is an update for the local player
	if moveType == playerMoveNoUpdate {
		if p.list[localPlayerID] == nil || p.list[localPlayerID].update == nil {
			// if the local player does have an update pending, we need to send an unchanged movement type instead
			bs.Clear()
			return
		}

		moveType = playerMoveUnchanged
	}

	// set the first bit to indicate we have an update
	bs.Set()

	// two bits represent the local player update type
	bs.SetBits(uint32(moveType), 2)

	switch moveType {
	case playerMoveUnchanged:
		// nothing to do

	case playerMoveWalk:
		// write 3 bits for the direction
		code := directionCodes[p.local.walkDirection]
		bs.SetBits(uint32(code), 3)

		// write 1 bit if a further update is required
		needsUpdate := p.list[localPlayerID].update != nil
		bs.SetOrClear(needsUpdate)

	case playerMoveRun:
		// TODO
		panic("not implemented")

	case playerMovePosition:
		// write 2 bits for the z coordinate
		bs.SetBits(uint32(p.local.position.Z), 2)

		// write 1 bit each for the clear waypoints and update needed flags
		needsUpdate := p.list[localPlayerID].update != nil
		bs.SetOrClear(p.local.clearWaypoints)
		bs.SetOrClear(needsUpdate)

		// write 7 bits each for the y and x coordinates
		bs.SetBits(uint32(p.local.position.Y), 7)
		bs.SetBits(uint32(p.local.position.X), 7)
	}
}

func (p *PlayerUpdateResponse) writeOtherMovements(bs *network.BitSet) {
	movements := 0
	for playerID, other := range p.list {
		if playerID == localPlayerID || other.movement == nil {
			continue
		}

		movements++
	}

	// write 8 bits indicating how many other players there are to update
	bs.SetBits(uint32(movements), 8)

	for playerID, other := range p.list {
		if playerID == localPlayerID || other.movement == nil {
			continue
		}

		// set or clear 1 bit to flag if an update is required or if this player should only be tracked
		moveType := other.movement.moveType
		if moveType == playerMoveNoUpdate {
			// if the player does have an update pending, we need to send an unchanged movement type instead
			if other.update == nil {
				bs.Clear()
				continue
			}

			moveType = playerMoveUnchanged
		}

		bs.Set()

		// write 2 bits for the movement type
		bs.SetBits(uint32(moveType), 2)

		switch moveType {
		case playerMoveUnchanged:
			// nothing to do

		case playerMoveWalk:
			// write 3 bits for the direction
			code := directionCodes[other.movement.walkDirection]
			bs.SetBits(uint32(code), 3)

			// write 1 bit if a further update is required
			// TODO: is this needed since the player is already in the player list?
			needsUpdate := p.list[playerID].update != nil
			bs.SetOrClear(needsUpdate)
			//bs.Clear()

		case playerMoveRun:
			// TODO
			panic("not implemented")
		}
	}
}

func (p *PlayerUpdateResponse) writePlayerList(playerIDs []int, bs *network.BitSet) {
	for _, playerID := range playerIDs {
		pl := p.list[playerID]

		// don't include the local player here or players with movements reported earlier
		if playerID == localPlayerID || pl.movement != nil {
			continue
		}

		// write 11 bits for the player id
		bs.SetBits(uint32(playerID), 11)

		// write 1 bit if the player is observed
		needsUpdate := pl.update != nil
		bs.SetOrClear(needsUpdate)

		// write 1 bit if the player should have their path waypoints cleared
		bs.SetOrClear(pl.clearWaypoints)

		// write 5 bits for the y and x coordinate offsets
		bs.SetBits(uint32(pl.pos.Y), 5)
		bs.SetBits(uint32(pl.pos.X), 5)
	}

	// mark the end of the player list
	bs.SetBits(0x7FF, 11)
}

func (p *PlayerUpdateResponse) writePlayerUpdates(playerIDs []int, w *network.ProtocolWriter) error {
	for _, playerID := range playerIDs {
		pl := p.list[playerID]

		// skip players with no pending updates
		if pl.update == nil {
			continue
		}

		err := p.writePlayerUpdate(pl.update, w)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PlayerUpdateResponse) writePlayerUpdate(update *playerUpdate, w *network.ProtocolWriter) error {
	// if the mask cannot fit into a single byte, split it into two
	if update.mask > 0xFF {
		err := w.WriteUint16(update.mask)
		if err != nil {
			return err
		}
	} else {
		err := w.WriteUint8(byte(update.mask))
		if err != nil {
			return err
		}
	}

	// write chat message
	if update.mask&updateChatMessageText != 0 {
		// write 2 bytes for the color and effect
		colorCode := common.ChatColorCode(update.chatMessage.Color)
		effectCode := common.ChatEffectCode(update.chatMessage.Effect)

		err := w.WriteUint16(uint16(effectCode)<<8 | uint16(colorCode))
		if err != nil {
			return err
		}

		// TODO: player rights
		// write 1 byte for the player rights of the sending player
		err = w.WriteUint8(0x00)
		if err != nil {
			return err
		}

		// encode the chat message
		encoded := util.EncodeChat(update.chatMessage.Text)
		reversed := make([]byte, len(encoded))
		for i := len(encoded) - 1; i >= 0; i-- {
			reversed[i] = encoded[len(encoded)-i-1]
		}

		// write 1 byte the length of the message (inverted)
		err = w.WriteUint8(byte(len(reversed) * -1))
		if err != nil {
			return err
		}

		// write the message text itself
		_, err = w.Write(reversed)
		if err != nil {
			return err
		}
	}

	// write appearance update
	if update.mask&updateAppearance != 0 {
		err := p.writeAppearance(update.appearance, w)
		if err != nil {
			return err
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
	err := bw.WriteUint8(gender)
	if err != nil {
		return err
	}

	// write 1 byte for overhead icon
	err = bw.WriteUint8(byte(a.OverheadIconID))
	if err != nil {
		return err
	}

	// write each equipment slot
	for i := 0; i < 12; i++ {
		appearanceID := p.appearanceIDForSlot(ea, i)

		// if nothing is present at this slot, write one byte only
		if appearanceID == 0 {
			err = bw.WriteUint8(0)
			if err != nil {
				return err
			}

			continue
		}

		// write 2 bytes for an equipped item or model id
		err = bw.WriteUint16(uint16(appearanceID))
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
	for _, color := range a.BodyColors {
		err = bw.WriteUint8(byte(color))
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

	// convert the name to a long integer and write it as 8 bytes
	name := util.EncodeName(ea.name)
	err = bw.WriteUint64(name)
	if err != nil {
		return err
	}

	// write a byte the combat level
	err = bw.WriteUint8(byte(a.CombatLevel))
	if err != nil {
		return err
	}

	// write 2 bytes for the skill level
	err = bw.WriteUint16(uint16(a.TotalLevel))
	if err != nil {
		return err
	}

	buffer, err := bw.Buffer()
	if err != nil {
		return err
	}

	// write a byte for the appearance buffer size. the client expects this to be a negative value.
	err = w.WriteUint8(byte(buffer.Len() * -1))
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

// appearanceIDForSlot returns the ID to use for an entity's appearance in a particular slot.
func (p *PlayerUpdateResponse) appearanceIDForSlot(ea *entityAppearance, slot int) int {
	a := ea.appearance

	switch slot {
	case 0:
		return a.Base.Head
	case 1:
		return a.Base.Face
	case 2:
		return p.equippedItemIDOrDefault(ea, model.EquipmentSlotTypeNecklace, 0)
	case 3:
		return p.equippedItemIDOrDefault(ea, model.EquipmentSlotTypeWeapon, 0)
	case 4:
		return p.equippedItemIDOrDefault(ea, model.EquipmentSlotTypeBody, a.Base.Body)
	case 5:
		return p.equippedItemIDOrDefault(ea, model.EquipmentSlotTypeShield, 0)
	case 6:
		return p.equippedItemIDOrDefault(ea, model.EquipmentSlotTypeCape, 0)
	case 7:
		return p.equippedItemIDOrDefault(ea, model.EquipmentSlotTypeLegs, a.Base.Legs)
	case 8:
		return a.Base.Arms
	case 9:
		return p.equippedItemIDOrDefault(ea, model.EquipmentSlotTypeHands, a.Base.Hands)
	case 10:
		return p.equippedItemIDOrDefault(ea, model.EquipmentSlotTypeFeet, a.Base.Feet)
	case 11:
		return p.equippedItemIDOrDefault(ea, model.EquipmentSlotTypeHead, 0)
	case 12:
		return p.equippedItemIDOrDefault(ea, model.EquipmentSlotTypeRing, 0)
	}

	return 0
}

// equippedItemIDOrDefault returns the item ID of an entity's equipment, or a default ID if nothing is equipped there.
func (p *PlayerUpdateResponse) equippedItemIDOrDefault(ea *entityAppearance, slotType model.EquipmentSlotType, i int) int {
	slot, ok := ea.appearance.Equipment[slotType]
	if ok {
		// offset the appearance id to differentiate it from the non-item ids
		return slot.Item.ID + 0x200
	}

	return i
}
