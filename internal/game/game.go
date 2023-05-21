package game

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/asset"
	"github.com/mbpolan/openmcs/internal/config"
	"github.com/mbpolan/openmcs/internal/interaction"
	"github.com/mbpolan/openmcs/internal/logger"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/network/response"
	"github.com/mbpolan/openmcs/internal/telemetry"
	"github.com/mbpolan/openmcs/internal/util"
	"github.com/pkg/errors"
	"strings"
	"sync"
	"time"
)

// playerMaxIdleInterval is the maximum time a player can be idle before being forcefully logged out.
// TODO: this should be configurable
const playerMaxIdleInterval = 3 * time.Minute

// tickInterval defines how long a single tick is.
const tickInterval = 600 * time.Millisecond

// itemDespawnInterval defines how long an Item remains on the map before despawning.
const itemDespawnInterval = 3 * time.Minute

// ErrConflict is reported when a player is already connected to the game.
var ErrConflict = errors.New("already logged in")

// Options are parameters that configure how the game engine behaves.
type Options struct {
	Config         *config.Config
	ItemAttributes []*model.ItemAttributes
	Telemetry      telemetry.Telemetry
}

// Game is the game engine and representation of the game world.
type Game struct {
	doneChan         chan bool
	interaction      *interaction.Manager
	interfaces       []*model.Interface
	items            map[int]*model.Item
	lastPlayerUpdate time.Time
	ticker           *time.Ticker
	mapManager       *MapManager
	mu               sync.RWMutex
	players          []*playerEntity
	objects          []*model.WorldObject
	playersOnline    sync.Map
	removePlayers    map[int]*playerEntity
	regions          map[model.Vector2D]*RegionManager
	scripts          *ScriptManager
	telemetry        telemetry.Telemetry
	tick             uint64
	welcomeMessage   string
	worldID          int
	worldMap         *model.Map
}

// NewGame creates a new game engine using the given configuration.
func NewGame(opts Options) (*Game, error) {
	g := &Game{
		doneChan:       make(chan bool, 1),
		interaction:    interaction.New(opts.Config.Interfaces),
		items:          map[int]*model.Item{},
		removePlayers:  map[int]*playerEntity{},
		telemetry:      opts.Telemetry,
		tick:           0,
		welcomeMessage: opts.Config.Server.WelcomeMessage,
		worldID:        opts.Config.Server.WorldID,
	}

	start := time.Now()

	// load scripts
	g.scripts = NewScriptManager("scripts", g)
	numScripts, err := g.scripts.Load()
	if err != nil {
		return nil, err
	}

	logger.Infof("loaded %d scripts in %s", numScripts, time.Now().Sub(start))
	start = time.Now()

	// load game assets
	err = g.loadAssets(opts.Config.Server.AssetDir, opts.ItemAttributes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load game asset")
	}

	logger.Infof("loaded assets in: %s", time.Now().Sub(start))
	start = time.Now()

	// initialize the map manager and perform a warm-up for map regions
	g.mapManager = NewMapManager(g.worldMap)
	g.mapManager.WarmUp()
	g.mapManager.Start()

	logger.Infof("finished map warm-up in: %s", time.Now().Sub(start))

	return g, nil
}

// Stop gracefully terminates the game loop.
func (g *Game) Stop() {
	g.ticker.Stop()
	g.mapManager.Stop()
	g.doneChan <- true
}

// Run starts the game loop and begins processing events. You need to call this method before players can connect
// and interact with the game.
func (g *Game) Run() {
	g.ticker = time.NewTicker(tickInterval)
	go g.loop()
}

// AddFriend attempts to add another player to the player's friends list.
func (g *Game) AddFriend(p *model.Player, username string) {
	g.addToList(p, username, true)
}

// RemoveFriend removes another player from the player's friends list.
func (g *Game) RemoveFriend(p *model.Player, username string) {
	g.removeFromList(p, username, true)
}

// AddIgnored attempts to add another player to the player's ignored list.
func (g *Game) AddIgnored(p *model.Player, username string) {
	g.addToList(p, username, false)
}

// RemoveIgnored removes another player from the player's ignored list.
func (g *Game) RemoveIgnored(p *model.Player, username string) {
	g.removeFromList(p, username, false)
}

// SetPlayerModes updates the chat and interaction modes for a player.
func (g *Game) SetPlayerModes(p *model.Player, publicChat model.ChatMode, privateChat model.ChatMode, interaction model.InteractionMode) {
	g.mu.Lock()
	pe := g.findPlayer(p)
	g.mu.Unlock()

	if pe == nil {
		return
	}

	// lock the player while we update their modes
	pe.mu.Lock()
	defer pe.mu.Unlock()

	// if this player's private chat mode has changed to be more or less restrictive, we need to broadcast their
	// online status again
	if pe.player.Modes.PrivateChat != privateChat {
		pe.MarkStatusBroadcast()
	}

	pe.player.Modes = model.PlayerModes{
		PublicChat:  publicChat,
		PrivateChat: privateChat,
		Interaction: interaction,
	}
}

// ProcessAbuseReport handles an abuse report sent by a player.
func (g *Game) ProcessAbuseReport(p *model.Player, username string, reason int, mute bool) {
	// TODO: log the report and mute target player, if they exist
	logger.Infof("player %s reported %s for abuse reason %d (mute? %t)", p.Username, username, reason, mute)
}

// MarkPlayerActive updates a player's last activity tracker and prevents them from becoming idle.
func (g *Game) MarkPlayerActive(p *model.Player) {
	pe, unlockFunc := g.findPlayerAndLockAll(p)
	defer unlockFunc()
	if pe == nil {
		return
	}

	pe.lastInteraction = time.Now()
}

// MarkPlayerInactive flags that a player's client reported them as being idle.
func (g *Game) MarkPlayerInactive(p *model.Player) {
	pe, unlockFunc := g.findPlayerAndLockAll(p)
	defer unlockFunc()
	if pe == nil {
		return
	}

	// TODO: still needed?
	//pe.scheduler.Plan(NewEventWithType(EventCheckIdleImmediate, time.Now()))
}

// DoInterfaceAction processes an action that a player performed on an interface.
func (g *Game) DoInterfaceAction(p *model.Player, interfaceID int) {
	pe, unlockFunc := g.findPlayerAndLockAll(p)
	unlockFunc()
	if pe == nil {
		return
	}

	pe.DeferDoInterfaceAction(interfaceID)
}

// DoInteractWithObject handles a player interaction with an object on the map.
func (g *Game) DoInteractWithObject(p *model.Player, action int, globalPos model.Vector2D) {
	// TODO
}

// DoSetPlayerDesign handles updating a player's character design.
func (g *Game) DoSetPlayerDesign(p *model.Player, gender model.EntityGender, base model.EntityBase, bodyColors []int) {
	pe := g.findPlayer(p)
	if pe == nil {
		return
	}

	pe.mu.Lock()
	defer pe.mu.Unlock()

	pe.player.Appearance.Gender = gender
	pe.player.Appearance.Base = base
	pe.player.Appearance.BodyColors = bodyColors
	pe.appearanceChanged = true
	pe.DeferHideInterfaces()
}

// DoPlayerChatCommand handles a chat command sent by a player.
func (g *Game) DoPlayerChatCommand(p *model.Player, text string) {
	pe, unlockFunc := g.findPlayerAndLockAll(p)
	defer unlockFunc()
	if pe == nil {
		return
	}

	// determine if a valid and recognized chat command was sent
	command := ParseChatCommand(text)
	if command == nil {
		return
	}

	g.handleChatCommand(pe, command)
}

// DoPlayerChat broadcasts a player's chat message to nearby players.
func (g *Game) DoPlayerChat(p *model.Player, effect model.ChatEffect, color model.ChatColor, text string) {
	pe, unlockFunc := g.findPlayerAndLockAll(p)
	defer unlockFunc()
	if pe == nil {
		return
	}

	pe.lastChatMessage = &model.ChatMessage{
		Color:  color,
		Effect: effect,
		Text:   text,
	}
	pe.lastChatTime = time.Now()
}

// DoPlayerPrivateChat sends a private chat message from one player to a recipient player.
func (g *Game) DoPlayerPrivateChat(p *model.Player, recipient string, text string) {
	// find the player, locking only the game
	pe, unlockFunc := g.findPlayerAndLockGame(p)
	if pe == nil {
		unlockFunc()
		return
	}

	// find the target player
	recipient = strings.ToLower(recipient)
	var target *playerEntity
	for _, other := range g.players {
		if strings.ToLower(other.player.Username) == recipient {
			target = other
			break
		}
	}

	// unlock the game at this point
	unlockFunc()

	// if the target player is not online, we don't need to do anything
	if target == nil {
		return
	}

	// lock both players and defer unlocking them until later
	pe.mu.Lock()
	target.mu.Lock()
	defer func() {
		pe.mu.Unlock()
		target.mu.Unlock()
	}()

	// if the target player has their private chat off or if it's in friends-only mode and this player is not on their
	// friends list, then don't send the message. also don't send messages to a player if the sending player is on
	// their ignored list.
	if target.player.Modes.PrivateChat == model.ChatModeOff || target.player.IsIgnored(pe.player.Username) ||
		(target.player.Modes.PrivateChat == model.ChatModeFriends && !(target.player.HasFriend(pe.player.Username))) {
		return
	}

	// all good; plan sending the message to the other player
	pm := response.NewPrivateChatResponse(target.privateMessageID, pe.player.Username, pe.player.Type, text)
	target.Send(pm)
	target.privateMessageID++
}

// WalkPlayer starts moving the player to a destination from a start position then following a set of waypoints. The
// slice of waypoints are deltas relative to start.
func (g *Game) WalkPlayer(p *model.Player, start model.Vector2D, waypoints []model.Vector2D) {
	pe, unlockFunc := g.findPlayerAndLockAll(p)
	defer unlockFunc()
	if pe == nil {
		return
	}

	// convert each waypoint into global coordinates
	actuals := []model.Vector2D{start}
	for _, w := range waypoints {
		actuals = append(actuals, model.Vector2D{
			X: start.X + w.X,
			Y: start.Y + w.Y,
		})
	}

	// start traversing from the player's current position
	from := pe.player.GlobalPos.To2D()
	var path []model.Vector2D

	// plan a direct path to each waypoint
	for _, w := range actuals {
		// find the distance from the previous position
		dx := w.X - from.X
		dy := w.Y - from.Y

		// waypoints may be many tiles away, so we need to plan each segment in the path individually
		for dx != 0 || dy != 0 {
			// prefer diagonal movement if the delta is only one tile away, followed by movement along the x-axis
			// anf lastly movement along the y-axis
			if dx != 0 && dy != 0 && util.Abs(dx) == util.Abs(dy) {
				path = append(path, model.Vector2D{
					X: from.X + util.Unit(dx),
					Y: from.Y + util.Unit(dy),
				})

				dx -= util.Unit(dx)
				dy -= util.Unit(dy)
			} else if dx != 0 {
				path = append(path, model.Vector2D{
					X: from.X + util.Unit(dx),
					Y: from.Y,
				})

				dx -= util.Unit(dx)
			} else if dy != 0 {
				path = append(path, model.Vector2D{
					X: from.X,
					Y: from.Y + util.Unit(dy),
				})

				dy -= util.Unit(dy)
			}

			// mark this position as the starting point for the next segment
			from = path[len(path)-1]
		}
	}

	logger.Debugf("path player %s via %+v", p.Username, path)
	pe.path = path
	pe.nextPathIdx = 0
}

// ValidatePlayer checks if a player can be added to the game.
func (g *Game) ValidatePlayer(p *model.Player) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// prevent the player from logging in again if they are already connected
	for _, tpe := range g.players {
		if tpe.player.ID == p.ID {
			return ErrConflict
		}
	}

	return nil
}

// AddPlayer joins a player to the world and handles ongoing game events and network interactions.
func (g *Game) AddPlayer(p *model.Player, writer *network.ProtocolWriter) {
	pe := newPlayerEntity(p, writer)

	// update the player's inventory and equipment to ensure items match their expected models. if an item does not
	// match, remove it from its respective location
	for _, slot := range pe.player.Inventory {
		if slot == nil {
			continue
		}

		item := g.items[slot.Item.ID]
		if item == nil {
			pe.player.ClearInventoryItem(slot.ID)
		} else {
			pe.player.SetInventoryItem(item, slot.Amount, slot.ID)
		}
	}

	for slotType, slot := range pe.player.Appearance.Equipment {
		item := g.items[slot.Item.ID]
		if item == nil {
			pe.player.ClearEquippedItem(slotType)
		} else {
			pe.player.SetEquippedItem(item, slot.Amount, slotType)
		}
	}

	// set initial client tab interface
	pe.tabInterfaces = g.interaction.ClientTabInterfaces(pe.player.EquippedWeaponStyle())

	// add the player to the player list
	g.mu.Lock()
	g.players = append(g.players, pe)
	g.mu.Unlock()

	// mark the player as being online and broadcast their status
	g.telemetry.RecordPlayerConnected()
	g.playersOnline.Store(pe.player.Username, true)
	pe.MarkStatusBroadcast()

	// start the player's event loop
	go g.playerLoop(pe)

	// compute the player's region and position
	regionOrigin, regionRelative := g.playerRegionPosition(pe)
	pe.regionOrigin = regionOrigin

	// plan an initial map region load
	region := response.NewLoadRegionResponse(regionOrigin)
	pe.Send(region)

	// plan an initial player update
	update := response.NewPlayerUpdateResponse(p.ID)
	update.SetLocalPlayerPosition(regionRelative, true)
	update.AddAppearanceUpdate(p.ID, p.Username, p.Appearance)
	pe.Send(update)

	// describe the local region
	// FIXME: this should be done in the game loop
	rg := util.RegionOriginToGlobal(regionOrigin)
	mapUpdates := g.mapManager.State(rg, model.BoundaryNone)
	if len(mapUpdates) > 0 {
		pe.Send(mapUpdates...)
	}

	// plan an initial character design if flagged
	if pe.player.UpdateDesign {
		pe.DeferShowInterface(g.interaction.CharacterDesigner.ID)
		pe.player.UpdateDesign = false
	}

	// plan an update to the client sidebar interface
	pe.DeferSendInterfaces()

	// plan an event to clear the player's equipment
	pe.DeferSendEquipment()

	// plan an event to clear the player's inventory
	pe.DeferSendInventory()

	// plan an update to the client's interaction modes
	pe.DeferSendModes()

	// plan an update to the player's skills
	pe.DeferSendSkills()

	// plan an update to the player's friends list
	pe.DeferSendFriendList()

	// plan an update to the player's ignored list
	pe.DeferSendIgnoreList()

	// plan a welcome message
	pe.DeferSendServerMessage(g.welcomeMessage)
}

// RemovePlayer removes a previously joined player from the world.
func (g *Game) RemovePlayer(p *model.Player) {
	pe, unlockFunc := g.findPlayerAndLockGame(p)
	defer unlockFunc()

	if pe == nil {
		return
	}

	g.handleRemovePlayer(pe)
}

// DoTakeGroundItem handles a player's request to pick up a ground item at a position, in global coordinates.
func (g *Game) DoTakeGroundItem(p *model.Player, itemID int, globalPos model.Vector2D) {
	// validate the item is known
	targetItem := g.items[itemID]
	if targetItem == nil {
		return
	}

	pe := g.findPlayer(p)
	if pe == nil {
		return
	}

	pe.mu.Lock()
	defer pe.mu.Unlock()

	// defer this action since the player might need to walk to the position of the item
	pe.DeferTakeGroundItemAction(targetItem, globalPos.To3D(pe.player.GlobalPos.Z))
}

// DoDropInventoryItem handles a player's request to drop an inventory item.
func (g *Game) DoDropInventoryItem(p *model.Player, itemID, interfaceID, secondaryActionID int) {
	// validate the item is known
	targetItem := g.items[itemID]
	if targetItem == nil {
		return
	}

	pe := g.findPlayer(p)
	if pe == nil {
		return
	}

	pe.mu.Lock()
	defer pe.mu.Unlock()

	// defer this action so it occurs on the next game tick
	pe.DeferDropInventoryItem(targetItem, interfaceID, secondaryActionID)
}

// DoSwapInventoryItem handles a player's request to move an item in their inventory to another slot.
func (g *Game) DoSwapInventoryItem(p *model.Player, fromSlot int, toSlot int, mode int) {
	pe := g.findPlayer(p)
	if pe == nil {
		return
	}

	// plan an update to the player's inventory slots
	pe.DeferMoveInventoryItem(fromSlot, toSlot)
}

// DoEquipItem handles a player's request to equip an item.
func (g *Game) DoEquipItem(p *model.Player, itemID, interfaceID, secondaryActionID int) {
	// validate the item is known
	targetItem := g.items[itemID]
	if targetItem == nil {
		return
	}

	pe := g.findPlayer(p)
	if pe == nil {
		return
	}

	pe.mu.Lock()
	defer pe.mu.Unlock()

	// defer the action to the next tick
	pe.DeferEquipItem(targetItem, interfaceID)
}

// DoUnequipItem handles a player's request to unequip an item.
func (g *Game) DoUnequipItem(p *model.Player, itemID, interfaceID int, slotType model.EquipmentSlotType) {
	// validate the item is known
	targetItem := g.items[itemID]
	if targetItem == nil {
		return
	}

	pe := g.findPlayer(p)
	if pe == nil {
		return
	}

	pe.mu.Lock()
	defer pe.mu.Unlock()

	// defer the action to the next tick
	pe.DeferUnequipItem(targetItem, interfaceID, slotType)
}

// DoUseItem handles a player's request to use an item.
func (g *Game) DoUseItem(p *model.Player, itemID, interfaceID, actionID int) {
	// TODO
}

// DoUseInventoryItem handles a player's request to use an inventory item on another item.
func (g *Game) DoUseInventoryItem(p *model.Player, sourceItemID, sourceInterfaceID, sourceSlotID,
	targetItemID, targetInterfaceID, targetSlotID int) {
	// TODO
}

// broadcastPlayerStatus sends updates to other players that have them on their friends lists. An optional list of
// target player usernames can be passed to limit who receives the update.
// Concurrency requirements: (a) game state should be locked and (b) all players should be locked.
func (g *Game) broadcastPlayerStatus(pe *playerEntity, targets ...string) {
	_, online := g.playersOnline.Load(pe.player.Username)

	// if this player has their private chat turned off, show them as offline to everyone
	if pe.player.Modes.PrivateChat == model.ChatModeOff {
		online = false
	}

	// find players that have this player on their friends list
	for _, other := range g.players {
		// skip the same player, or if they are not targeted
		if pe == other || (len(targets) > 0 && !util.Contains(targets, other.player.Username)) {
			continue
		}

		// check if the other player has this player on their friends list, and that this player does not have the
		// other on their ignored list
		if other.player.HasFriend(pe.player.Username) {
			// if the player's private chat mode is friends-only, then we only show them online if the two are mutual
			// friends and if the other player is not on their ignored list
			onlineForOther := online
			if pe.player.Modes.PrivateChat == model.ChatModeFriends && !pe.player.HasFriend(other.player.Username) ||
				pe.player.IsIgnored(other.player.Username) {
				onlineForOther = false
			}

			var update *response.FriendStatusResponse
			if onlineForOther {
				update = response.NewFriendStatusResponse(pe.player.Username, g.worldID)
			} else {
				update = response.NewOfflineFriendStatusResponse(pe.player.Username)
			}

			err := update.Write(other.writer)
			if err != nil {
				// other player could have disconnected, so we don't treat this as a fatal error
				logger.Debugf("failed to send friend status update for %s to %s: %s", pe.player.Username, other.player.Username, err)
			}
		}
	}
}

// findSpectators returns a slice of players that are within visual distance of a given player.
// Concurrency requirements: (a) game state should be locked and (b) all players should be locked.
func (g *Game) findSpectators(pe *playerEntity) []*playerEntity {
	var others []*playerEntity
	for _, tpe := range g.players {
		// ignore our own player and others players on different z coordinates
		if tpe.player.ID == pe.player.ID || tpe.player.GlobalPos.Z != pe.player.GlobalPos.Z {
			continue
		}

		// compute their distance to the player and add them as a spectator if they are within range
		// TODO: make this configurable?
		dx := util.Abs(tpe.player.GlobalPos.X - pe.player.GlobalPos.X)
		dy := util.Abs(tpe.player.GlobalPos.Y - pe.player.GlobalPos.Y)
		if dx <= 14 && dy <= 14 {
			others = append(others, tpe)
		}
	}

	return others
}

// findPlayerAndLockGame returns the playerEntity for the corresponding player, locking only the game mutex. You must
// call the returned function to properly unlock the mutex.
// Concurrency requirements: (a) game state should NOT be locked and (b) all players should NOT be locked.
func (g *Game) findPlayerAndLockGame(p *model.Player) (*playerEntity, func()) {
	g.mu.RLock()

	pe := g.findPlayer(p)
	if pe == nil {
		g.mu.RUnlock()

		return nil, func() {}
	}

	return pe, func() {
		g.mu.RUnlock()
	}
}

// findPlayerAndLockAll returns the playerEntity for the corresponding player, locking the game and playerEntity mutexes
// along the way. You must call the returned function to properly unlock all mutexes.
// Concurrency requirements: (a) game state should NOT be locked and (b) all players should NOT be locked.
func (g *Game) findPlayerAndLockAll(p *model.Player) (*playerEntity, func()) {
	g.mu.RLock()

	pe := g.findPlayer(p)
	if pe == nil {
		g.mu.RUnlock()

		return nil, func() {}
	}

	pe.mu.Lock()

	return pe, func() {
		pe.mu.Unlock()
		g.mu.RUnlock()
	}
}

// findPlayer returns the playerEntity for the corresponding player. This method does not lock; if you need thread
// safety, use the findPlayerAndLockAll method instead.
// Concurrency requirements: (a) game state should be locked and (b) all players should not be locked.
func (g *Game) findPlayer(p *model.Player) *playerEntity {
	var tpe *playerEntity
	for _, pe := range g.players {
		if pe.player.ID == p.ID {
			tpe = pe
			break
		}
	}

	return tpe
}

// loop continuously runs the main game server update cycle.
func (g *Game) loop() {
	for {
		select {
		case <-g.doneChan:
			logger.Infof("stopping game engine")
			return
		case <-g.ticker.C:
			start := time.Now()

			err := g.handleGameUpdate()
			if err != nil {
				logger.Errorf("ending game state update due to error: %s", err)
				return
			}

			// record how long this game state update took
			end := time.Now().Sub(start)
			g.telemetry.RecordGameStateUpdateDuration(float64(end.Nanoseconds()))
		}
	}
}

// playerLoop continuously monitors for player-bound events.
// Concurrency requirements: (a) game state should NOT be locked and (b) this player should NOT be locked.
func (g *Game) playerLoop(pe *playerEntity) {
	for {
		select {
		case <-pe.doneChan:
			// send a graceful disconnect to the client if possible
			err := pe.writer.WriteUint8(response.DisconnectResponseHeader)
			if err != nil {
				logger.Debugf("failed to write disconnect response to player %d: %s", pe.player.ID, err)
			}

			return

		case update := <-pe.outChan:
			// send a response to the player
			err := update.Write(pe.writer)
			if err != nil {
				logger.Errorf("ending player loop due to error on update: %s", err)
				return
			}
		}
	}
}

// loadAssets reads and parses all game asset.
// Concurrency requirements: none (any locks may be held).
func (g *Game) loadAssets(assetDir string, itemAttributes []*model.ItemAttributes) error {
	var err error
	manager := asset.NewManager(assetDir)

	// load interfaces
	g.interfaces, err = manager.Interfaces()
	if err != nil {
		return err
	}

	// load world objects
	g.objects, err = manager.WorldObjects()
	if err != nil {
		return err
	}

	// load map data
	g.worldMap, err = manager.Map(g.objects)
	if err != nil {
		return err
	}

	// load items
	items, err := manager.Items()
	if err != nil {
		return err
	}

	// create a map of item ids to their models
	for _, item := range items {
		g.items[item.ID] = item
	}

	// assign item attributes to items
	for _, attr := range itemAttributes {
		item, ok := g.items[attr.ItemID]
		if !ok {
			logger.Warnf("item attribute does not match any known item with ID: %d", attr.ItemID)
			continue
		}

		item.Attributes = attr
	}

	return nil
}

// findEffectiveRegion computes the region origin, in region coordinates, that the player's client should render.
func (g *Game) findEffectiveRegion(pe *playerEntity) model.Vector2D {
	regionGlobal := util.GlobalToRegionGlobal(pe.player.GlobalPos)
	base := util.RegionGlobalToClientBase(regionGlobal)

	// compute the ending bounds of the area the client knows about, relative to the client base coordinates
	baseBound := model.Vector2D{
		X: base.X + util.ClientChunkArea2D.X*util.Chunk2D.X,
		Y: base.Y + util.ClientChunkArea2D.Y*util.Chunk2D.Y,
	}

	regionOrigin := util.GlobalToRegionOrigin(regionGlobal).To2D()

	// determine if the player is nearing or has encroached on the boundary to a new map region. if so, adjust the
	// region origin so that it matches the newly discovered region.
	if util.Abs(pe.player.GlobalPos.X-base.X) < util.RegionBoundary2D.X {
		regionOrigin.X -= util.Chunk2D.X
	} else if util.Abs(pe.player.GlobalPos.X-baseBound.X) < util.RegionBoundary2D.X {
		regionOrigin.X += util.Chunk2D.X
	} else if util.Abs(pe.player.GlobalPos.Y-base.Y) < util.RegionBoundary2D.Y {
		regionOrigin.Y -= util.Chunk2D.Y
	} else if util.Abs(pe.player.GlobalPos.Y-baseBound.Y) < util.RegionBoundary2D.Y {
		regionOrigin.Y += util.Chunk2D.Y
	}

	return regionOrigin
}

// playerRegionPosition returns the region origin and player position relative to that origin.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) playerRegionPosition(pe *playerEntity) (model.Vector2D, model.Vector3D) {
	// find the region where the player's client is rendered
	regionOrigin := g.findEffectiveRegion(pe)

	// compute the current region origin in global coordinates and the client base
	regionGlobal := util.RegionOriginToGlobal(regionOrigin)
	base := util.RegionGlobalToClientBase(regionGlobal)

	// compute the player's position relative to the client's base coordinates. the client uses a top-left origin
	// which is offset by six chunks on each axis.
	regionRelative := model.Vector3D{
		X: pe.player.GlobalPos.X - base.X,
		Y: pe.player.GlobalPos.Y - base.Y,
		Z: pe.player.GlobalPos.Z,
	}

	return util.GlobalToRegionOrigin(regionGlobal).To2D(), regionRelative
}

// handleRemovePlayer adds a player to the list of players that will be removed from the game.
// Concurrency requirements: (a) game state should be locked and (b) this player may be locked.
func (g *Game) handleRemovePlayer(pe *playerEntity) {
	// add this player to the removal list, and let the next state update actually remove them
	g.removePlayers[pe.player.ID] = pe
}

// handleChatCommand processes a chat command sent by a player.
// Concurrency requirements: (a) game state should NOT be locked and (b) this player should NOT be locked.
func (g *Game) handleChatCommand(pe *playerEntity, command *ChatCommand) {
	switch command.Type {
	case ChatCommandTypeSpawnItem:
		params := command.SpawnItem

		// prevent invalid items from being spawned
		if _, ok := g.items[params.ItemID]; ok {
			g.mapManager.AddGroundItem(params.ItemID, params.Amount, params.Amount > 1, params.DespawnTimeSeconds, pe.player.GlobalPos)
		} else {
			pe.Send(response.NewServerMessageResponse(fmt.Sprintf("Invalid item: %d", command.SpawnItem.ItemID)))
		}

	case ChatCommandTypeClearTile:
		// remove all ground items at player's position
		g.mapManager.ClearGroundItems(pe.player.GlobalPos)

	case ChatCommandTeleportRelative:
		// relocate the player to a new location relative to their current position
		newPos := pe.player.GlobalPos.Add(command.Pos)
		pe.teleportGlobal = &newPos

	case ChatCommandTeleport:
		// relocate the player to a new location
		pe.teleportGlobal = &command.Pos

	case ChatCommandTypePosition:
		// send a message containing player's server position on the world map
		msg := fmt.Sprintf("GlobalPos: %d, %d, %d", pe.player.GlobalPos.X, pe.player.GlobalPos.Y, pe.player.GlobalPos.Z)
		pe.Send(response.NewServerMessageResponse(msg))

	case ChatCommandCharacterDesigner:
		// open the character designer interface
		pe.DeferShowInterface(g.interaction.CharacterDesigner.ID)

	case ChatCommandShowInterface:
		// show an interface
		pe.DeferShowInterface(command.ShowInterface.InterfaceID)

	case ChatCommandHideInterfaces:
		// clear all interfaces
		pe.DeferHideInterfaces()
	}
}

// addToList adds another player to the player's friends or ignore list.
// Concurrency requirements: (a) game state should NOT be locked and (b) all players should NOT be locked.
func (g *Game) addToList(p *model.Player, username string, friend bool) {
	target := strings.Trim(strings.ToLower(username), " ")

	// TODO: validate if target player exists in persistent storage and get their properly cased name

	// TODO: update list in persistent storage

	// we need to manually control the player's lock
	pe, unlockFunc := g.findPlayerAndLockGame(p)
	defer unlockFunc()
	if pe == nil {
		return
	}

	pe.mu.Lock()
	defer pe.mu.Unlock()

	// is this player already in their friend's list
	exists := false
	if friend {
		exists = pe.player.HasFriend(target)
	} else {
		exists = pe.player.IsIgnored(target)
	}

	// avoid adding duplicates
	if exists {
		return
	}

	if friend {
		pe.player.Friends = append(pe.player.Friends, target)
	} else {
		pe.player.Ignored = append(pe.player.Ignored, target)
	}

	// find the target player, if they are online
	var targetPlayer *playerEntity
	for _, tpe := range g.players {
		if strings.ToLower(tpe.player.Username) == target {
			targetPlayer = tpe
			break
		}
	}

	// send a status update about the player that was just added and vice versa
	pe.MarkStatusBroadcastTarget(target)
	if targetPlayer != nil {
		pe.MarkStatusBroadcastTarget(pe.player.Username)
	}
}

// removeFromList removes another player from the player's friends or ignored list. No mutexes should be locked when
// calling this method.
// Concurrency requirements: (a) game state should NOT be locked and (b) all players should NOT be locked.
func (g *Game) removeFromList(p *model.Player, username string, friend bool) {
	// only lock the game state temporarily
	g.mu.Lock()
	pe := g.findPlayer(p)
	g.mu.Unlock()

	if pe == nil {
		return
	}

	pe.mu.Lock()
	defer pe.mu.Unlock()

	// remove the player from the appropriate list
	var list []string
	if friend {
		list = pe.player.Friends
	} else {
		list = pe.player.Ignored
	}

	target := strings.Trim(strings.ToLower(username), " ")
	for i, other := range list {
		if strings.ToLower(other) == target {
			list = append(list[:i], list[i+1:]...)

			// if the player's private chat is set to friends-only, we need to broadcast their status in case the removed
			// player should no longer see them online
			pe.MarkStatusBroadcastTarget(target)
			break
		}
	}

	// the client automatically removes players, so we don't need to send an explicit list update
	if friend {
		pe.player.Friends = list
	} else {
		pe.player.Ignored = list
	}
}

// addPlayerInventoryItem adds an item to the player's inventory, if there is room, and plans an update to the player's
// client. The item may or may not be stackable.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) addPlayerInventoryItem(pe *playerEntity, item *model.Item, amount int) {
	slotID := -1
	totalAmount := amount

	// stackable items occupy the same slot, so we need to handle them separately
	if item.Stackable {
		// find an existing slot that has this item, if one exists. if there isn't one, fall through and treat this
		// stackable as a new item
		slot := pe.player.InventorySlotWithItem(item.ID)
		if slot != nil {
			// can this slot accommodate the additional stack amount? if not, the player cannot hold this item
			if int64(slot.Amount+amount) > model.MaxStackableSize {
				return
			}

			slotID = slot.ID
			totalAmount = slot.Amount + amount
		}
	}

	// find the next available slot in the player's inventory, if one exists
	if slotID == -1 {
		slotID = pe.player.NextFreeInventorySlot()
	}

	// if there is no available slot, the player cannot hold this item
	if slotID == -1 {
		return
	}

	// set the item on the slot
	pe.player.SetInventoryItem(item, totalAmount, slotID)

	// update the player's inventory
	inventory := response.NewSetInventoryItemResponse(g.interaction.InventoryTab.SlotsID)
	inventory.AddSlot(slotID, item.ID, totalAmount)
	pe.Send(inventory)
}

// dropPlayerInventoryItem removes the first occurrence of an item from the player's inventory, and adds it to the
// world map.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) dropPlayerInventoryItem(pe *playerEntity, item *model.Item) *model.InventorySlot {
	slot := pe.player.InventorySlotWithItem(item.ID)
	if slot == nil {
		return nil
	}

	// remove the item from the player's inventory
	pe.player.ClearInventoryItem(slot.ID)

	// update the player's inventory
	inventory := response.NewSetInventoryItemResponse(g.interaction.InventoryTab.SlotsID)
	inventory.ClearSlot(slot.ID)
	pe.Send(inventory)

	return slot
}

// equipPlayerInventoryItem removes the first occurrence of an item in the player's inventory and adds it to their
// currently equipped item set. If an item of the same slot type is already equipped, the two will be swapped.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) equipPlayerInventoryItem(pe *playerEntity, item *model.Item) {
	if !item.CanEquip() {
		return
	}

	// find the first instance of the item in the player's inventory
	invSlot := pe.player.InventorySlotWithItem(item.ID)
	if invSlot == nil {
		return
	}

	// prepare an update for the player's inventory
	inventory := response.NewSetInventoryItemResponse(g.interaction.InventoryTab.SlotsID)

	// find the target equipment slot. if there is already an item equipped, we need to either swap the two or add to
	// the equipped item's stack if the item is stackable
	amount := int64(invSlot.Amount)
	equipSlot := pe.player.EquipmentSlot(item.Attributes.EquipSlotType)
	if equipSlot != nil {
		if item.Stackable {
			// add as much to the equipped item's stack as possible. if the amount in the inventory exceeds the total
			// stack size, take only as much as can be stacked
			amount = int64(equipSlot.Amount) + int64(amount)
			if amount > model.MaxStackableSize {
				amount = model.MaxStackableSize
				remaining := int(amount - model.MaxStackableSize)

				// leave the remaining amount in the inventory
				pe.player.SetInventoryItem(invSlot.Item, remaining, invSlot.ID)
				inventory.AddSlot(invSlot.ID, invSlot.Item.ID, remaining)
			} else {
				// since there is capacity to add the entire inventory stack into the equipped item's stack, we can
				// clear the inventory slot
				pe.player.ClearInventoryItem(invSlot.ID)
				inventory.ClearSlot(invSlot.ID)
			}
		} else {
			// put the equipped item into the same inventory slot as the incoming item
			pe.player.SetInventoryItem(equipSlot.Item, equipSlot.Amount, invSlot.ID)
			inventory.AddSlot(invSlot.ID, equipSlot.Item.ID, equipSlot.Amount)
		}
	} else {
		// remove the item from its inventory slot
		pe.player.ClearInventoryItem(invSlot.ID)
		inventory.ClearSlot(invSlot.ID)
	}

	// equip the item into the slot
	pe.player.SetEquippedItem(item, int(amount), item.Attributes.EquipSlotType)

	// update the player's inventory
	pe.Send(inventory)

	// update the player's equipment status
	equipment := response.NewSetInventoryItemResponse(g.interaction.EquipmentTab.SlotsID)
	equipment.AddSlot(int(item.Attributes.EquipSlotType), invSlot.Item.ID, invSlot.Amount)
	pe.Send(equipment)

	// update the player's equipment interface and their equipped weapon interface tabs
	pe.Send(g.interaction.EquipmentTab.Update(pe.player)...)
	pe.Send(response.NewSidebarInterfaceResponse(model.ClientTabEquippedItem,
		g.interaction.WeaponTab.IDForWeaponStyle(pe.player.EquippedWeaponStyle())))

	// mark that we need to update the player's appearance if necessary
	if item.Attributes.EquipSlotType.Visible() {
		pe.appearanceChanged = true
	}
}

// unequipPlayerInventoryItem removes an equipped item and places it in the player's inventory. The player should have
// room in their inventory prior to calling this method.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) unequipPlayerInventoryItem(pe *playerEntity, item *model.Item, slotType model.EquipmentSlotType) {
	// validate the player still has this item equipped
	slot := pe.player.EquipmentSlot(slotType)
	if slot == nil {
		return
	}

	// remove the item from the player's equipment
	pe.player.ClearEquippedItem(slotType)

	// add it to their inventory
	g.addPlayerInventoryItem(pe, slot.Item, slot.Amount)

	// update the player's equipment status
	equipment := response.NewSetInventoryItemResponse(g.interaction.EquipmentTab.SlotsID)
	equipment.ClearSlot(int(item.Attributes.EquipSlotType))
	pe.Send(equipment)

	// update the player's equipment interface and their equipped weapon interface tabs
	pe.Send(g.interaction.EquipmentTab.Update(pe.player)...)
	pe.Send(response.NewSidebarInterfaceResponse(model.ClientTabEquippedItem,
		g.interaction.WeaponTab.IDForWeaponStyle(pe.player.EquippedWeaponStyle())))

	// mark that we need to update the player's appearance if necessary
	if slotType.Visible() {
		pe.appearanceChanged = true
	}
}

// handleGameUpdate performs a game state update.
// Concurrency requirements: (a) game state should NOT be locked and (b) all players should NOT be locked.
func (g *Game) handleGameUpdate() error {
	g.mu.Lock()

	// lock all players and check for inactive players that should be disconnected
	for _, pe := range g.players {
		pe.mu.Lock()

		// add this player to the removal list if they have idled for too long
		if time.Now().Sub(pe.lastInteraction) >= playerMaxIdleInterval {
			g.removePlayers[pe.player.ID] = pe
		}
	}

	// remove and disconnected players first
	for _, pe := range g.removePlayers {
		idx := -1

		// remove this player from the tracking lists of other players
		for i, tpe := range g.players {
			if tpe.player.ID == pe.player.ID {
				idx = i
			} else {
				delete(tpe.tracking, pe.player.ID)
			}
		}

		if idx > -1 {
			// drop the player from the player list
			g.players = append(g.players[:idx], g.players[idx+1:]...)

			// mark the player as offline and broadcast their status
			g.telemetry.RecordPlayerDisconnected()
			g.playersOnline.Delete(pe.player.Username)
			g.broadcastPlayerStatus(pe)

			// flag the player's event loop to stop and disconnect them
			pe.mu.Unlock()
			pe.Drop()
		}
	}

	g.removePlayers = map[int]*playerEntity{}
	changedAppearance := map[int]bool{}
	changedRegions := map[int]bool{}

	// process each player, handling deferred actions and movement sequences
	for _, pe := range g.players {
		// prepare a new player update or use the pending, existing one
		if pe.nextUpdate == nil {
			pe.nextUpdate = response.NewPlayerUpdateResponse(pe.player.ID)
		}
		update := pe.nextUpdate

		// has this player teleported to a new location?
		if pe.teleportGlobal != nil {
			pe.player.GlobalPos = *pe.teleportGlobal
			origin, relative := g.playerRegionPosition(pe)

			if origin != pe.regionOrigin {
				pe.regionOrigin = origin

				region := response.NewLoadRegionResponse(pe.regionOrigin)
				pe.Send(region)
			}

			relocate := response.NewPlayerUpdateResponse(pe.player.ID)
			relocate.SetLocalPlayerPosition(relative, true)
			pe.Send(relocate)

			changedRegions[pe.player.ID] = true
			pe.teleportGlobal = nil
		}

		// handle a deferred action for the player
		g.handleDeferredActions(pe)

		// if this player's appearance has changed, we need to include it in their update
		if pe.appearanceChanged {
			update.AddAppearanceUpdate(pe.player.ID, pe.player.Username, pe.player.Appearance)
			pe.appearanceChanged = false
			changedAppearance[pe.player.ID] = true
		}

		// check if the player is walking
		if pe.Walking() {
			next := pe.path[pe.nextPathIdx]

			// add the change in direction to the local player's movement
			dir := model.DirectionFromDelta(next.Sub(pe.player.GlobalPos.To2D()))
			update.SetLocalPlayerWalk(dir)

			// update the player's position
			pe.player.GlobalPos.X = next.X
			pe.player.GlobalPos.Y = next.Y

			// move past this path segment
			pe.nextPathIdx++

			// check if the player has moved into a new map region, and schedule a map region load is that's the case
			origin := g.findEffectiveRegion(pe)
			if origin != pe.regionOrigin {
				region := response.NewLoadRegionResponse(origin)
				pe.Send(region)

				// determine in which direction the new region is in relative to the player's current region. use this
				// to apply a trimming boundary so that we don't send map state for an overlapping part of the two
				// regions
				delta := pe.regionOrigin.Sub(origin)
				boundary := model.BoundaryNone
				if delta.X < 0 {
					boundary |= model.BoundaryWest
				} else if delta.X > 0 {
					boundary |= model.BoundaryEast
				}

				if delta.Y < 0 {
					boundary |= model.BoundarySouth
				} else if delta.Y > 0 {
					boundary |= model.BoundaryNorth
				}

				// send the map state for the new region
				state := g.mapManager.State(util.RegionOriginToGlobal(origin), boundary)
				if len(state) > 0 {
					pe.Send(state...)
				}

				// mark this as the current region the player's client has loaded
				pe.regionOrigin = origin
				changedRegions[pe.player.ID] = true
			}
		}

		// broadcast this player's status to friends and other target players if required
		if pe.nextStatusBroadcast != nil {
			g.broadcastPlayerStatus(pe, pe.nextStatusBroadcast.targets...)
			pe.nextStatusBroadcast = nil
		}
	}

	// reconcile the map state now that players have taken their actions
	mapUpdates := g.mapManager.Reconcile()

	// update each player with actions and movements of those nearby
	for _, pe := range g.players {
		update := pe.nextUpdate

		// is this player in a region that has map updates? only send updates if they have not left this region
		regionGlobal := util.RegionOriginToGlobal(g.findEffectiveRegion(pe))
		if updates, ok := mapUpdates[regionGlobal]; ok && !changedRegions[pe.player.ID] {
			pe.Send(updates...)
		}

		// find players within visual distance of this player
		others := g.findSpectators(pe)
		for _, other := range others {
			_, known := pe.tracking[other.player.ID]

			// determine what to do with this player. there are several possibilities:
			// (a) this is the first time we've seen them: send an update including their appearance and location
			// (b) we've seen them before, but we don't yet have their update captured
			// (c) we've seen them before and know their last update: no action needed
			if !known {
				posOffset := other.player.GlobalPos.Sub(pe.player.GlobalPos).To2D()
				update.AddToPlayerList(other.player.ID, posOffset, true, true)

				update.AddAppearanceUpdate(other.player.ID, other.player.Username, other.player.Appearance)
				pe.tracking[other.player.ID] = other
			} else {
				theirUpdate := other.nextUpdate

				// if the other player does not have an update, do not change their posture relative to us. otherwise
				// synchronize with their local movement
				if theirUpdate == nil {
					update.AddOtherPlayerNoUpdate(other.player.ID)
				} else {
					update.SyncLocalMovement(other.player.ID, theirUpdate)
				}

				// if the player has changed their appearance, include it in the update
				if changedAppearance[other.player.ID] {
					update.AddAppearanceUpdate(other.player.ID, other.player.Username, other.player.Appearance)
				}
			}

			// if the other player has posted a chat message, determine if this player should receive it
			if other.lastChatTime.After(pe.chatHighWater) && other.lastChatMessage != nil {
				// only receive messages if the other player has not ignored the sending player
				receive := !other.player.IsIgnored(pe.player.Username)

				switch pe.player.Modes.PublicChat {
				case model.ChatModePublic, model.ChatModeHide:
					// receive all chat messages, allowing client to hide chat messages on demand

				case model.ChatModeFriends:
					// TODO: check if the other player is a friend

				case model.ChatModeOff:
					// do not receive any messages
					receive = false
				}

				// only include this chat message if the player should receive it
				if receive {
					update.AddChatMessage(other.player.ID, other.lastChatMessage)
				}
			}
		}

		pe.chatHighWater = time.Now()
	}

	// unlock all players and dispatch their updates
	for _, pe := range g.players {
		pe.mu.Unlock()
		pe.outChan <- pe.nextUpdate
		pe.nextUpdate = nil
	}

	g.mu.Unlock()
	return nil
}

// handleDeferredActions processes scheduled actions for a player.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleDeferredActions(pe *playerEntity) {
	deferredActions := pe.TickDeferredActions()
	for _, deferred := range deferredActions {
		switch deferred.ActionType {
		case ActionSendServerMessage:
			g.handleServerMessage(pe, deferred.ServerMessageAction.Message)
			pe.RemoveDeferredAction(deferred)

		case ActionMoveInventoryItem:
			g.handlePlayerSwapInventoryItem(pe, deferred.MoveInventoryItemAction)
			pe.RemoveDeferredAction(deferred)

		case ActionSendSkills:
			g.handleSendPlayerSkills(pe)
			pe.RemoveDeferredAction(deferred)

		case ActionSendInterfaces:
			g.handleSendPlayerInterfaces(pe)
			pe.RemoveDeferredAction(deferred)

		case ActionSendModes:
			g.handleSendPlayerModes(pe)
			pe.RemoveDeferredAction(deferred)

		case ActionSendFriendList:
			g.handleSendPlayerFriendList(pe)
			pe.RemoveDeferredAction(deferred)

		case ActionSendIgnoreList:
			g.handleSendPlayerIgnoreList(pe)
			pe.RemoveDeferredAction(deferred)

		case ActionSendEquipment:
			g.handleSendPlayerEquipment(pe)
			pe.Send(g.interaction.EquipmentTab.Update(pe.player)...)
			pe.RemoveDeferredAction(deferred)

		case ActionSendInventory:
			g.handleSendPlayerInventory(pe)
			pe.RemoveDeferredAction(deferred)

		case ActionTakeGroundItem:
			action := deferred.TakeGroundItem

			// pick up a ground item only if the player has reached the position of that item
			if pe.player.GlobalPos != action.GlobalPos {
				break
			}

			// check if the player has room in their inventory
			if !pe.player.InventoryCanHoldItem(action.Item) {
				pe.Send(response.NewServerMessageResponse("You cannot carry any more items"))
				pe.RemoveDeferredAction(deferred)
				break
			}

			// remove the ground item if it still exists, and allow the next reconciliation to take care of
			// updating the state of the map
			item := g.mapManager.RemoveGroundItem(action.Item.ID, action.GlobalPos)
			if item != nil {
				// add the item to the player's inventory
				g.addPlayerInventoryItem(pe, action.Item, item.Amount)
			}

			pe.RemoveDeferredAction(deferred)

		case ActionDropInventoryItem:
			action := deferred.DropInventoryItemAction

			// remove the item from the player's inventory
			slot := g.dropPlayerInventoryItem(pe, action.Item)
			if slot != nil {
				// put the item on the tile the player's standing on, and let the next reconciliation take care of
				// updating the state of the map
				timeout := int(itemDespawnInterval.Seconds())
				g.mapManager.AddGroundItem(slot.Item.ID, slot.Amount, action.Item.Stackable, &timeout, pe.player.GlobalPos)
			}

			pe.RemoveDeferredAction(deferred)

		case ActionEquipItem:
			action := deferred.EquipItemAction

			g.equipPlayerInventoryItem(pe, action.Item)
			pe.RemoveDeferredAction(deferred)

		case ActionUnequipItem:
			action := deferred.UnequipItemAction

			// validate that the player has room in their inventory
			if !pe.player.InventoryCanHoldItem(action.Item) {
				pe.Send(response.NewServerMessageResponse("You have no room for this item in your inventory"))
				pe.RemoveDeferredAction(deferred)
				break
			}

			g.unequipPlayerInventoryItem(pe, action.Item, action.SlotType)
			pe.RemoveDeferredAction(deferred)

		case ActionShowInterface:
			action := deferred.ShowInterfaceAction

			pe.Send(response.NewShowInterfaceResponse(action.InterfaceID))
			pe.RemoveDeferredAction(deferred)

		case ActionHideInterfaces:
			pe.Send(&response.ClearScreenResponse{})
			pe.RemoveDeferredAction(deferred)

		case ActionDoInterfaceAction:
			action := deferred.DoInterfaceAction

			err := g.scripts.DoInterface(pe, action.InterfaceID, 0)
			if err != nil {
				logger.Warnf("failed to execute interface %d script: %s", action.InterfaceID, err)
			}

			pe.RemoveDeferredAction(deferred)

		default:
		}
	}
}

// handleSendPlayerSkills handles sending a player their current skill levels and experience.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleSendPlayerSkills(pe *playerEntity) {
	var responses []response.Response

	for _, skill := range pe.player.Skills {
		responses = append(responses, response.NewSkillDataResponse(skill))
	}

	pe.Send(responses...)
}

// handleSendPlayerEquipment handles sending a player their current equipped items.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleSendPlayerEquipment(pe *playerEntity) {
	equipment := response.NewSetInventoryItemResponse(g.interaction.EquipmentTab.SlotsID)
	for _, slotType := range model.EquipmentSlotTypes {
		slot := pe.player.EquipmentSlot(slotType)
		if slot == nil {
			equipment.ClearSlot(int(slotType))
		} else {
			equipment.AddSlot(int(slotType), slot.Item.ID, slot.Amount)
		}
	}

	pe.Send(equipment)
}

// handleSendPlayerEquipment handles sending a player their current inventory items.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleSendPlayerInventory(pe *playerEntity) {
	inventory := response.NewSetInventoryItemResponse(g.interaction.InventoryTab.SlotsID)
	for id, slot := range pe.player.Inventory {
		if slot == nil {
			inventory.ClearSlot(id)
		} else {
			inventory.AddSlot(slot.ID, slot.Item.ID, slot.Amount)
		}
	}

	pe.Send(inventory)
}

// handleSendPlayerInterfaces handles sending a player the tab interface their client should display.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleSendPlayerInterfaces(pe *playerEntity) {
	var responses []response.Response

	for _, tab := range model.ClientTabs {
		// does the player have an interface for this tab?
		var r *response.SidebarInterfaceResponse
		if id, ok := pe.tabInterfaces[tab]; ok {
			r = response.NewSidebarInterfaceResponse(tab, id)
		} else {
			r = response.NewRemoveSidebarInterfaceResponse(tab)
		}

		responses = append(responses, r)
	}

	pe.Send(responses...)
}

// handleSendPlayerModes handles sending a player their current chat modes.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleSendPlayerModes(pe *playerEntity) {
	modes := response.NewSetModesResponse(pe.player.Modes.PublicChat, pe.player.Modes.PrivateChat, pe.player.Modes.Interaction)
	pe.Send(modes)
}

// handleSendPlayerFriendList handles sending a player their friends list and the status of each friend.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleSendPlayerFriendList(pe *playerEntity) {
	// tell the client the list is loading
	status := response.NewFriendsListStatusResponse(model.FriendsListStatusLoading)
	pe.Send(status)

	// send the status for each friends list entry
	var responses []response.Response
	for _, username := range pe.player.Friends {
		var friend *response.FriendStatusResponse

		// send a status for this player if they are online or not
		_, ok := g.playersOnline.Load(username)
		if ok {
			friend = response.NewFriendStatusResponse(username, 69)
		} else {
			friend = response.NewOfflineFriendStatusResponse(username)
		}

		responses = append(responses, friend)
	}

	// finally, tell the client the list has been sent
	status = response.NewFriendsListStatusResponse(model.FriendsListStatusLoaded)
	responses = append(responses, status)
	pe.Send(responses...)
}

// handleSendPlayerIgnoreList handles sending a player their ignore list.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleSendPlayerIgnoreList(pe *playerEntity) {
	ignored := response.NewIgnoredListResponse(pe.player.Ignored)
	pe.Send(ignored)
}

// handleServerMessage handles sending a player a server message.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleServerMessage(pe *playerEntity, message string) {
	msg := response.NewServerMessageResponse(message)
	pe.Send(msg)
}

// handlePlayerSwapInventoryItem handles moving an item from one slot to another in a player's inventory.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handlePlayerSwapInventoryItem(pe *playerEntity, action *MoveInventoryItemAction) {
	inventory := response.NewSetInventoryItemResponse(g.interaction.InventoryTab.SlotsID)

	// make sure there is still an item at the starting slot
	fromSlot := pe.player.Inventory[action.FromSlot]
	if fromSlot == nil {
		return
	}

	// if there is already an item at the target slot, move it to the starting slot. otherwise clear the item at
	// the starting slot
	toSlot := pe.player.Inventory[action.ToSlot]
	if toSlot != nil {
		pe.player.SetInventoryItem(toSlot.Item, toSlot.Amount, fromSlot.ID)
		inventory.AddSlot(fromSlot.ID, toSlot.Item.ID, toSlot.Amount)
	} else {
		pe.player.ClearInventoryItem(action.FromSlot)
		inventory.ClearSlot(action.FromSlot)
	}

	// move the item from the starting slot to the target slot
	pe.player.SetInventoryItem(fromSlot.Item, fromSlot.Amount, action.ToSlot)
	inventory.AddSlot(action.ToSlot, fromSlot.Item.ID, fromSlot.Amount)

	pe.Send(inventory)
}
