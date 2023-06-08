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
	"math"
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

// maxPlayers is the maximum amount of players that can be connected to the game server.
const maxPlayers = 2000

// maxSkillExperience is the maximum amount of experience a player can have in a skill.
const maxSkillExperience = 200_000_000

// ValidationResult enumerates the errors that can result from game engine player validation.
type ValidationResult int

const (
	// ValidationResultSuccess indicates no error was found during player validation.
	ValidationResultSuccess ValidationResult = iota
	// ValidationResultAlreadyLoggedIn indicates a player already has an existing session.
	ValidationResultAlreadyLoggedIn
	// ValidationResultNoCapacity indicates the server cannot accommodate more players.
	ValidationResultNoCapacity
)

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
	interfaces       map[int]*model.Interface
	items            map[int]*model.Item
	lastPlayerUpdate time.Time
	ticker           *time.Ticker
	mapManager       *MapManager
	mu               sync.RWMutex
	players          []*playerEntity
	playerIndices    [maxPlayers]int
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
		interfaces:     map[int]*model.Interface{},
		items:          map[int]*model.Item{},
		playerIndices:  [maxPlayers]int{},
		removePlayers:  map[int]*playerEntity{},
		telemetry:      opts.Telemetry,
		tick:           0,
		welcomeMessage: opts.Config.Server.WelcomeMessage,
		worldID:        opts.Config.Server.WorldID,
	}

	// initialize player index tracker
	for i := range g.playerIndices {
		g.playerIndices[i] = -1
	}

	start := time.Now()

	// load scripts
	g.scripts = NewScriptManager(opts.Config.Server.ScriptsDir, g)
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
	pe := g.findPlayer(p)
	if pe == nil {
		return
	}

	// validate the interface is known
	actor, ok := g.interfaces[interfaceID]
	if !ok {
		return
	}

	// find the parent interface that should receive the action
	parent := actor
	for parent.Parent != nil {
		parent = parent.Parent
	}

	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.DeferDoInterfaceAction(parent, actor)
}

// DoInteractWithObject handles a player interaction with an object on the map.
func (g *Game) DoInteractWithObject(p *model.Player, action int, globalPos model.Vector2D) {
	// TODO
}

// DoCastSpellOnItem handles a player casting a spell on one of their inventory items.
func (g *Game) DoCastSpellOnItem(p *model.Player, slotID, itemID, inventoryInterfaceID, spellInterfaceID int) {
	pe := g.findPlayer(p)
	if pe == nil {
		return
	}

	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.DeferCastSpellOnItem(slotID, itemID, inventoryInterfaceID, spellInterfaceID)
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

	// require the player to have administrator privileges before executing a command
	if pe.player.Type != model.PlayerAdmin {
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
func (g *Game) ValidatePlayer(p *model.Player) ValidationResult {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// check if the server is at capacity
	if len(g.players) == maxPlayers {
		return ValidationResultNoCapacity
	}

	// prevent the player from logging in again if they are already connected
	for _, tpe := range g.players {
		if tpe.player.ID == p.ID {
			return ValidationResultAlreadyLoggedIn
		}
	}

	return ValidationResultSuccess
}

// AddPlayer joins a player to the world and handles ongoing game events and network interactions. The lowMemory flag
// indicates if the player opted to play in low-memory mode on the client.
func (g *Game) AddPlayer(p *model.Player, lowMemory bool, writer *network.ProtocolWriter) {
	pe := newPlayerEntity(p, writer)
	pe.isLowMemory = lowMemory

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

	// add the player to the player list, and assign them their index on the server player list
	g.mu.Lock()
	for i, used := range g.playerIndices {
		if used == -1 {
			g.playerIndices[i] = pe.player.ID
			pe.index = i
			break
		}
	}
	g.players = append(g.players, pe)
	g.mu.Unlock()

	// mark the player as being online and broadcast their status
	g.telemetry.RecordPlayerConnected()
	g.playersOnline.Store(pe.player.Username, true)
	pe.MarkStatusBroadcast()

	// start the player's event loop
	go g.playerLoop(pe)

	// tell the player's client that the player is initialized on the server
	pe.Send(&response.InitPlayerResponse{
		Member:      pe.player.Member,
		ServerIndex: pe.index,
	})

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
	g.checkScript(g.scripts.DoPlayerInit(pe))

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

	// plan an update to the player's run energy
	pe.DeferSendRunEnergy()

	// plan an update to the player's weight
	pe.DeferSendWeight()

	// reset the music track after log in
	pe.DeferPlayMusic(response.PlayMusicNoneID)

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

// checkScript validates and logs the result of a script execution.
func (g *Game) checkScript(err error) {
	// TODO: find a way to track this?
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
	interfaces, err := manager.Interfaces()
	if err != nil {
		return err
	}

	// create a map of interface ids to their models
	for _, inf := range interfaces {
		g.interfaces[inf.ID] = inf
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
	}

	if util.Abs(pe.player.GlobalPos.Y-base.Y) < util.RegionBoundary2D.Y {
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

// handleSetSidebarInterface sends a player's client an interface to show on a sidebar tab.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleSetSidebarInterface(pe *playerEntity, interfaceID, sidebarID int) {
	tab := model.ClientTab(sidebarID)

	pe.tabInterfaces[tab] = interfaceID
	pe.Send(response.NewSidebarInterfaceResponse(tab, interfaceID))
}

// handleClearSidebarInterface sends a player's client a command to remove an interface on a sidebar tab.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleClearSidebarInterface(pe *playerEntity, sidebarID int) {
	tab := model.ClientTab(sidebarID)

	delete(pe.tabInterfaces, tab)
	pe.Send(response.NewRemoveSidebarInterfaceResponse(model.ClientTab(sidebarID)))
}

// handleSetInterfaceColor sends a player's client a color to set for an interface.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleSetInterfaceColor(pe *playerEntity, interfaceID int, color model.Color) {
	r := &response.SetInterfaceColorResponse{
		InterfaceID: interfaceID,
		Color:       color,
	}

	pe.Send(r)
}

// handleSetInterfaceModel sends a player's client an item model to show on an interface.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleSetInterfaceModel(pe *playerEntity, interfaceID, itemID, zoom int) {
	r := &response.SetInterfaceModelResponse{
		InterfaceID: interfaceID,
		ItemID:      itemID,
		Zoom:        zoom,
	}

	pe.Send(r)
}

// handleSetInterfaceText sends a player's client the text to show on an interface.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleSetInterfaceText(pe *playerEntity, interfaceID int, text string) {
	pe.Send(response.NewSetInterfaceTextResponse(interfaceID, text))
}

// handleSetInterfaceSetting sends a setting value for the current interface.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleSetInterfaceSetting(pe *playerEntity, settingID, value int) {
	r := &response.SetInterfaceSettingResponse{
		SettingID: settingID,
		Value:     value,
	}

	pe.Send(r)
}

// handleRemovePlayer adds a player to the list of players that will be removed from the game.
// Concurrency requirements: (a) game state should be locked and (b) this player may be locked.
func (g *Game) handleRemovePlayer(pe *playerEntity) {
	// add this player to the removal list, and let the next state update actually remove them
	g.removePlayers[pe.player.ID] = pe
}

// handleConsumeInventoryItems attempts to consume a set of items from the player's inventory, returning true if
// successful or false if not. itemIDAmounts should be a vararg slice consisting of the item ID followed by the amount.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleConsumeInventoryItems(pe *playerEntity, itemIDAmounts ...int) bool {
	itemTargetSlots := map[int]*model.InventorySlot{}
	itemAmounts := map[int]int{}
	consumedSlots := map[int]bool{}

	// find the inventory slots which contain the necessary items with minimum amounts
	for i := 0; i < len(itemIDAmounts); i += 2 {
		itemID := itemIDAmounts[i]
		amount := itemIDAmounts[i+1]
		itemAmounts[itemID] = amount

		// validate the item is known
		item, ok := g.items[itemID]
		if !ok {
			return false
		}

		// find a candidate slot in the inventory
		ok = false
		for _, slot := range pe.player.Inventory {
			// skip empty slots or slots that we've already marked as consumed
			if slot == nil || consumedSlots[slot.ID] {
				continue
			}

			// check if this slot contains the item we're looking for. if the item is stackable, see if the amount in
			// this stack is sufficient. if the item is not stackable, we just need to check if this slot contains
			// it. once we find a candidate, mark the slot as consumed so we don't revisit it again
			if slot.Item.ID == itemID && (!item.Stackable || (item.Stackable && slot.Amount >= amount)) {
				itemTargetSlots[itemID] = slot
				consumedSlots[slot.ID] = true
				ok = true
				break
			}
		}

		// fail fast if the player doesn't meet the requirements for this item
		if !ok {
			return false
		}
	}

	// deduct and/or remove items from inventory
	inventory := response.NewSetInventoryItemResponse(g.interaction.InventoryTab.SlotsID)
	for itemID, slot := range itemTargetSlots {
		g.consumePlayerInventorySlot(pe, g.items[itemID], slot, itemAmounts[itemID], inventory)
	}

	// update the player's inventory now that we're done
	pe.Send(inventory)
	return true
}

// handleConsumeInventoryItemInSlot attempts to consume an item at a particular slot in the player's inventory,
// returning true if successful or false if not.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleConsumeInventoryItemInSlot(pe *playerEntity, slotID, amount int) bool {
	slot := pe.player.Inventory[slotID]
	if slot == nil {
		return false
	}

	// consume the item at the specified inventory slot
	inventory := response.NewSetInventoryItemResponse(g.interaction.InventoryTab.SlotsID)
	g.consumePlayerInventorySlot(pe, slot.Item, slot, amount, inventory)
	pe.Send(inventory)
	return true
}

// handleAddInventoryItem adds an item with an amount to the player's inventory. If the player's inventory is full,
// the item is dropped on the ground instead.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleAddInventoryItem(pe *playerEntity, itemID, amount int) {
	item, ok := g.items[itemID]
	if !ok {
		return
	}

	// if there is no room in the player's inventory, drop the item instead
	if !pe.player.InventoryCanHoldItem(item) {
		timeout := int(itemDespawnInterval.Seconds())
		g.mapManager.AddGroundItem(item.ID, amount, item.Stackable, &timeout, pe.player.GlobalPos)
		return
	}

	g.addPlayerInventoryItem(pe, item, amount)
}

// handleCountInventoryItems returns the number of items that existing in the player's inventory. If an item is
// stackable, the total number of stacked items of that kind will be returned. Otherwise, a count of each instance
// of a non-stackable item will be returned,
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleCountInventoryItems(pe *playerEntity, itemID int) int {
	count := 0
	for _, slot := range pe.player.Inventory {
		if slot == nil || slot.Item.ID != itemID {
			continue
		}

		count += slot.Amount
	}

	return count
}

// handleSendServerMessage sends a server message to a player.
// Concurrency requirements: (a) game state may be locked and (b) this player may be locked.
func (g *Game) handleSendServerMessage(pe *playerEntity, message string) {
	pe.Send(response.NewServerMessageResponse(message))
}

// handleChatCommand processes a chat command sent by a player.
// Concurrency requirements: (a) game state should be locked and (b) this player should NOT be locked.
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
		pe.DeferTeleportPlayer(newPos)

	case ChatCommandTeleport:
		// relocate the player to a new location
		pe.DeferTeleportPlayer(command.Pos)

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

	case ChatCommandReloadScripts:
		// clear script manager and reload all scripts
		// FIXME: this probably should be scheduled until _after_ the next tick
		n, err := g.scripts.Load()
		if err != nil {
			logger.Errorf("failed to reload scripts via command: %s", err)
		} else {
			logger.Infof("reloaded %d scripts via command", n)
		}

	case ChatCommandAnimate:
		// the player requested an animation
		pe.SetAnimation(command.Animate.ID, -1)
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
// client. The item may or may not be stackable. The player's weight will be updated after the fact.
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

	// update the player's weight
	g.sendPlayerWeight(pe)
}

// dropPlayerInventoryItem removes the first occurrence of an item from the player's inventory, and adds it to the
// world map. The player's weight will be updated after the fact.
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

	// update the player's weight
	g.sendPlayerWeight(pe)

	return slot
}

// consumeItemInSlot consumes an item in an inventory slot. If an item is stackable, its stack amount will be decreased.
// If th item is not stackable, it will be removed entirely. Response r will be modified to reflect the new state of
// the inventory slot. The player's weight will be updated after the fact.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) consumePlayerInventorySlot(pe *playerEntity, item *model.Item, slot *model.InventorySlot, amount int,
	r *response.SetInventoryItemsResponse) {

	// if the item is stackable, deduct from its stack and remove the item if the stack is then empty
	if item.Stackable {
		slot.Amount -= amount

		// if this slot is now empty, remove it entirely
		if slot.Amount == 0 {
			pe.player.ClearInventoryItem(slot.ID)
			r.ClearSlot(slot.ID)
		} else {
			r.AddSlot(slot.ID, item.ID, slot.Amount)
		}
	} else {
		pe.player.ClearInventoryItem(slot.ID)
		r.ClearSlot(slot.ID)
	}

	g.sendPlayerWeight(pe)
}

// equipPlayerInventoryItem removes the first occurrence of an item in the player's inventory and adds it to their
// currently equipped item set. If an item of the same slot type is already equipped, the two will be swapped.
// The player's weight will be updated after the fact.
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
	equipment := response.NewSetInventoryItemResponse(g.interaction.EquipmentTab.ID)
	equipment.AddSlot(int(item.Attributes.EquipSlotType), invSlot.Item.ID, invSlot.Amount)
	pe.Send(equipment)

	// update the player's weight
	g.sendPlayerWeight(pe)

	// update the player's equipment interface and their equipped weapon interface tabs
	g.checkScript(g.scripts.DoOnEquipItem(pe, item))

	// mark that we need to update the player's appearance if necessary
	if item.Attributes.EquipSlotType.Visible() {
		pe.appearanceChanged = true
	}
}

// unequipPlayerInventoryItem removes an equipped item and places it in the player's inventory. The player should have
// room in their inventory prior to calling this method. The player's weight will be updated after the fact.
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
	equipment := response.NewSetInventoryItemResponse(g.interaction.EquipmentTab.ID)
	equipment.ClearSlot(int(item.Attributes.EquipSlotType))
	pe.Send(equipment)

	// update the player's weight
	g.sendPlayerWeight(pe)

	// update the player's equipment interface and their equipped weapon interface tabs
	g.checkScript(g.scripts.DoOnUnequipItem(pe, item))

	// mark that we need to update the player's appearance if necessary
	if slotType.Visible() {
		pe.appearanceChanged = true
	}
}

// sendPlayerWeight sends the player their current weight.
func (g *Game) sendPlayerWeight(pe *playerEntity) {
	weight := &response.PlayerWeightResponse{Weight: int(pe.player.Weight())}
	pe.Send(weight)
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

	// remove any disconnected players first
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
			g.playerIndices[pe.index] = -1

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
	changedRegions := map[int]bool{}

	// process each player, handling deferred actions and movement sequences
	for _, pe := range g.players {
		// prepare a new player update or use the pending, existing one
		if pe.nextUpdate == nil {
			pe.nextUpdate = response.NewPlayerUpdateResponse(pe.index)
		}
		update := pe.nextUpdate

		// handle a deferred action for the player
		result := g.handleDeferredActions(pe)
		if result&ActionResultChangeRegions != 0 {
			changedRegions[pe.player.ID] = true
		}

		// add the player's current animation if there is one in progress. otherwise, if the current animation should
		// be completed, we need to explicitly tell the client to reset it.
		// this needs to be done before we check for appearance changes, since animations require another appearance
		// update be sent if they changed.
		if pe.Animating() {
			update.AddAnimation(pe.index, pe.AnimationID(), 0)

			// if the animation has an expiration, decrement the tick count and clear it if needed
			if pe.animationTicks > -1 {
				pe.animationTicks--
				if pe.animationTicks == 0 {
					pe.ClearAnimation()
					update.ClearAnimation(pe.index)
				}
			}
		} else if result&ActionResultClearAnimations != 0 {
			update.ClearAnimation(pe.index)
		}

		// do the same for the player's graphic, if one is currently set
		if pe.HasGraphic() {
			if !pe.graphicApplied {
				update.AddGraphic(pe.index, pe.GraphicID(), pe.GraphicHeight(), pe.GraphicDelay())
				pe.graphicApplied = true
			}

			// if the graphic has an expiration, decrement the tick count and clear it if needed
			if pe.graphicTicks > -1 {
				pe.graphicTicks--
				if pe.graphicTicks == 0 {
					pe.ClearGraphic()
					update.ClearGraphic(pe.index)
				}
			}
		} else if result&ActionResultClearGraphics != 0 {
			update.ClearGraphic(pe.index)
		}

		// if this player's appearance has changed, we need to include it in their update
		if pe.appearanceChanged {
			update.AddAppearanceUpdate(pe.index, pe.player.Username, pe.player.Appearance)
			pe.appearanceChanged = false
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
				update.AddToPlayerList(other.index, posOffset, true, true)

				update.AddAppearanceUpdate(other.index, other.player.Username, other.player.Appearance)
				update.ClearAnimation(other.index)
				pe.tracking[other.player.ID] = other
			} else {
				theirUpdate := other.nextUpdate

				// if the other player does not have an update, do not change their posture relative to us. otherwise
				// synchronize with their local movement
				if theirUpdate == nil {
					update.AddOtherPlayerNoUpdate(other.index)
				} else {
					update.SyncLocalMovement(other.index, theirUpdate)
					update.SyncLocalUpdate(other.index, theirUpdate)
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
					update.AddChatMessage(other.index, other.lastChatMessage)
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
// Concurrency requirements: (a) game state should be locked and (b) this player should be locked.
func (g *Game) handleDeferredActions(pe *playerEntity) ActionResult {
	result := ActionResultNoChange

	deferredActions := pe.TickDeferredActions()
	for _, deferred := range deferredActions {
		switch deferred.ActionType {
		case ActionDelayCurrent:
			// nothing to do; this action simply blocks other actions from being processed until it expires
			pe.RemoveDeferredAction(deferred)

		case ActionSendServerMessage:
			g.handleServerMessage(pe, deferred.ServerMessageAction.Message)
			pe.RemoveDeferredAction(deferred)

		case ActionMoveInventoryItem:
			g.handlePlayerSwapInventoryItem(pe, deferred.MoveInventoryItemAction)
			pe.RemoveDeferredAction(deferred)

		case ActionSendSkills:
			g.handleSendPlayerSkills(pe)
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
			pe.RemoveDeferredAction(deferred)

		case ActionSendInventory:
			g.handleSendPlayerInventory(pe)
			pe.RemoveDeferredAction(deferred)

		case ActionTakeGroundItem:
			action := deferred.TakeGroundItem

			// pick up a ground item only if the player has reached the position of that item
			if pe.player.GlobalPos != action.GlobalPos {
				result = ActionResultPending
				return result
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

			// execute a script for the parent interface
			err := g.scripts.DoInterface(pe, action.Parent, action.Actor, 0)
			if err != nil {
				logger.Warnf("failed to execute interface %d script: %s", action.Parent.ID, err)
			}

			pe.RemoveDeferredAction(deferred)

		case ActionTeleportPlayer:
			action := deferred.TeleportPlayerAction

			// move the player to the new position
			pe.player.GlobalPos = action.GlobalPos
			origin, relative := g.playerRegionPosition(pe)

			// remove this player from the tracking list of other players
			for _, tpe := range g.players {
				if tpe.player.ID != pe.player.ID {
					delete(tpe.tracking, pe.player.ID)
				}
			}
			pe.tracking = map[int]*playerEntity{}

			// if the teleport position is in another region, we need to send a region change
			if origin != pe.regionOrigin {
				pe.regionOrigin = origin

				result |= ActionResultChangeRegions
				region := response.NewLoadRegionResponse(pe.regionOrigin)
				pe.Send(region)
			}

			pe.nextUpdate.SetLocalPlayerPosition(relative, true)
			pe.nextUpdate.AddAppearanceUpdate(pe.index, pe.player.Username, pe.player.Appearance)
			pe.RemoveDeferredAction(deferred)

		case ActionCastSpellOnItem:
			action := deferred.CastSpellOnItemAction

			// find the parent spell book interface and inventory interfaces
			spellInterface := g.interfaces[action.SpellInterfaceID]
			if spellInterface == nil {
				pe.RemoveDeferredAction(deferred)
				break
			}

			spellBookInterface := spellInterface.Parent
			inventoryInterface := g.interfaces[action.InventoryInterfaceID]

			// find the slot containing the item
			slot := pe.player.Inventory[action.SlotID]

			// validate all interfaces exists, and the slot and item exist
			if inventoryInterface == nil || spellBookInterface == nil || slot == nil || slot.Item.ID != action.ItemID {
				pe.RemoveDeferredAction(deferred)
				break
			}

			// execute a script to handle this spell
			done, err := g.scripts.DoCastSpellOnItem(pe, slot.Item, action.SlotID, inventoryInterface, spellBookInterface, spellInterface)
			if err != nil {
				logger.Warnf("failed to execute cast item on spell script: %s", err)
			}

			// remove this deferred action, and bail out early if the spell has scheduled more blocking actions
			pe.RemoveDeferredAction(deferred)
			if !done {
				return ActionResultPending
			}

		case ActionExperienceGrant:
			action := deferred.ExperienceGrantAction

			// grant the player experience in the given skill
			experience := math.Min(pe.player.SkillExperience(action.SkillType)+action.Experience, maxSkillExperience)
			pe.player.SetSkillExperience(action.SkillType, experience)

			// send the client an experience drop update
			skillData := response.NewSkillDataResponse(pe.player.Skills[action.SkillType])
			pe.Send(skillData)

			pe.RemoveDeferredAction(deferred)

		case ActionSendRunEnergy:
			energy := &response.PlayerRunEnergyResponse{RunEnergy: int(pe.player.RunEnergy)}
			pe.Send(energy)
			pe.RemoveDeferredAction(deferred)

		case ActionSendWeight:
			g.sendPlayerWeight(pe)
			pe.RemoveDeferredAction(deferred)

		case ActionPlayMusic:
			action := deferred.PlayMusicAction

			music := &response.PlayMusicResponse{MusicID: action.MusicID}
			pe.Send(music)
			pe.RemoveDeferredAction(deferred)

		default:
		}
	}

	return result
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
	equipment := response.NewSetInventoryItemResponse(g.interaction.EquipmentTab.ID)
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

// handleTeleportPlayer teleports a player to another location.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleTeleportPlayer(pe *playerEntity, globalPos model.Vector3D) {
	// defer the teleport action for its tick delay
	pe.DeferTeleportPlayer(globalPos)
}

// handleAnimatePlayer sets a player's current animation with an expiration after a number of game ticks.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleAnimatePlayer(pe *playerEntity, animationID, tickDuration int) {
	pe.SetAnimation(animationID, tickDuration)
}

// handleSetPlayerGraphic sets a graphic to display with the player model at a height offset from the ground. A
// client-side tick delay can be provided to delay the start of the graphic being applied, and an expiration after a
// number of game ticks when the graphic will be removed.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleSetPlayerGraphic(pe *playerEntity, graphicID, height, delay, tickDuration int) {
	pe.SetGraphic(graphicID, height, delay, tickDuration)
}

// handleGrantExperience grants a player experience points in a skill.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleGrantExperience(pe *playerEntity, skillType model.SkillType, experience float64) {
	pe.DeferExperienceGrant(skillType, experience)
}

// handleSetSidebarTab sets the active tab on the client's sidebar.
// Concurrency requirements: (a) game state may be locked and (b) this player may be locked.
func (g *Game) handleSetSidebarTab(pe *playerEntity, tab model.ClientTab) {
	resp := &response.SidebarTabResponse{TabID: tab}
	pe.Send(resp)
}

// handleChangePlayerMovementSpeed changes the movement speed of a player.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleChangePlayerMovementSpeed(pe *playerEntity, speed model.MovementSpeed) {
	// TODO: check run energy and other preconditions
	pe.movementSpeed = speed
}

// handleChangePlayerAutoRetaliate changes a player's auto-retaliate combat option.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleChangePlayerAutoRetaliate(pe *playerEntity, enabled bool) {
	// TODO: if we're in combat this setting should take effect the next turn
	pe.player.AutoRetaliate = enabled
}

// handleSetPlayerQuestStatus updates the status of a quest for a player.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleSetPlayerQuestStatus(pe *playerEntity, questID int, status model.QuestStatus) {
	pe.player.SetQuestStatus(questID, status)
}

// handleSetPlayerQuestFlag sets a quest flag with a value for a player.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleSetPlayerQuestFlag(pe *playerEntity, questID, flagID, value int) {
	pe.player.SetQuestFlag(questID, flagID, value)
}

// handleSetPlayerMusicTrackUnlocked sets a music track as (un)locked for a player.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleSetPlayerMusicTrackUnlocked(pe *playerEntity, musicID int, enabled bool) {
	pe.player.SetMusicTrackUnlocked(musicID, enabled)
}

// handlePlayMusic sends the player's client a music track to play.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handlePlayMusic(pe *playerEntity, musicID int) {
	pe.DeferPlayMusic(musicID)
}

// handleShowInterface shows an interface for a player.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) handleShowInterface(pe *playerEntity, interfaceID int) {
	pe.DeferShowInterface(interfaceID)
}

// handleDelayCurrentAction blocks the player from performing other actions until a set amount of game ticks have
// elapsed.
func (g *Game) handleDelayCurrentAction(pe *playerEntity, tickDuration int) {
	pe.DeferActionCompletion(tickDuration)
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
