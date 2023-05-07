package game

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/asset"
	"github.com/mbpolan/openmcs/internal/config"
	"github.com/mbpolan/openmcs/internal/logger"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/network/response"
	"github.com/mbpolan/openmcs/internal/util"
	"github.com/pkg/errors"
	"strings"
	"sync"
	"time"
)

// playerMaxIdleInterval is the maximum time a player can be idle before being forcefully logged out.
// TODO: this should be configurable
const playerMaxIdleInterval = 3 * time.Minute

// playerUpdateInterval defines how often player updates are sent.
const playerUpdateInterval = 200 * time.Millisecond

// playerWalkInterval defines how long to wait between a player walks to their next waypoint.
const playerWalkInterval = 600 * time.Millisecond

// itemDespawnInterval defines how long an item remains on the map before despawning.
const itemDespawnInterval = 3 * time.Minute

// ErrConflict is reported when a player is already connected to the game.
var ErrConflict = errors.New("already logged in")

// Game is the game engine and representation of the game world.
type Game struct {
	items            []*model.Item
	doneChan         chan bool
	ticker           *time.Ticker
	objects          []*model.WorldObject
	players          []*playerEntity
	playersOnline    sync.Map
	lastPlayerUpdate time.Time
	worldMap         *model.Map
	mu               sync.RWMutex
	welcomeMessage   string
	removePlayers    []*playerEntity
	regions          map[model.Vector2D]*RegionManager
	worldID          int
	mapManager       *MapManager
}

// NewGame creates a new game engine using the given configuration.
func NewGame(cfg *config.Config) (*Game, error) {
	g := &Game{
		doneChan:       make(chan bool, 1),
		welcomeMessage: cfg.Server.WelcomeMessage,
		worldID:        cfg.Server.WorldID,
	}

	start := time.Now()

	// load game assets
	err := g.loadAssets(cfg.Server.AssetDir)
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
	g.ticker = time.NewTicker(50 * time.Millisecond)
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

	pe.scheduler.Plan(NewEventWithType(EventCheckIdleImmediate, time.Now()))
}

// DoInterfaceAction processes an action that a player performed on an interface.
func (g *Game) DoInterfaceAction(p *model.Player, action int) {
	pe, unlockFunc := g.findPlayerAndLockAll(p)
	unlockFunc()
	if pe == nil {
		return
	}

	// TODO: these should be scriptable
	// action on client logout button
	if action == 2458 {
		// TODO: check if player can be logged out (ie: are they in combat, etc.)
		_ = g.disconnect(pe)
	}
}

// DoInteractWithObject handles a player interaction with an object on the map.
func (g *Game) DoInteractWithObject(p *model.Player, action int, globalPos model.Vector2D) {
	// TODO
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
	target.PlanEvent(NewSendResponseEvent(pm, time.Now()))
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
	pe.lastWalkTime = time.Now().Add(-1 * playerWalkInterval)
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

	// set initial client tab interfaces
	// TODO: these ids should not be hardcoded
	pe.tabInterfaces = map[model.ClientTab]int{
		model.ClientTabSkills:      3917,
		model.ClientTabInventory:   3213,
		model.ClientTabFriendsList: 5065,
		model.ClientTabIgnoreList:  5715,
		model.ClientTabLogout:      2449,
	}

	g.mu.Lock()

	// add the player to the player list
	g.players = append(g.players, pe)

	g.mu.Unlock()

	// mark the player as being online and broadcast their status
	g.playersOnline.Store(pe.player.Username, true)
	pe.MarkStatusBroadcast()

	// start the player's event loop
	go g.playerLoop(pe)

	// compute the player's region and position
	regionOrigin, regionRelative := g.playerRegionPosition(pe)
	pe.regionOrigin = regionOrigin

	// plan an initial map region load
	region := response.NewLoadRegionResponse(regionOrigin)
	pe.PlanEvent(NewSendResponseEvent(region, time.Now()))

	// plan an initial player update
	update := response.NewPlayerUpdateResponse(p.ID)
	update.SetLocalPlayerPosition(regionRelative, true)
	update.AddAppearanceUpdate(p.ID, p.Username, p.Appearance)
	pe.PlanEvent(NewSendResponseEvent(update, time.Now()))

	// describe the local region
	// FIXME: this should be done in the game loop
	rg := util.RegionOriginToGlobal(regionOrigin)
	mapUpdates := g.mapManager.State(rg, model.BoundaryNone)
	if len(mapUpdates) > 0 {
		pe.PlanEvent(NewSendMultipleResponsesEvent(mapUpdates, time.Now()))
	}

	// plan an update to the client sidebar interfaces
	pe.PlanEvent(NewEventWithType(EventUpdateTabInterfaces, time.Now()))

	// plan an event to clear the player's inventory
	// TODO: the interface id should not be hardcoded
	inventory := response.NewClearInventoryResponse(3214)
	pe.PlanEvent(NewSendResponseEvent(inventory, time.Now()))

	// plan an update to the client's interaction modes
	modes := response.NewSetModesResponse(pe.player.Modes.PublicChat, pe.player.Modes.PrivateChat, pe.player.Modes.Interaction)
	pe.PlanEvent(NewSendResponseEvent(modes, time.Now()))

	// plan an update to the player's skills
	pe.PlanEvent(NewEventWithType(EventSkills, time.Now()))

	// plan an update to the player's friends list
	pe.PlanEvent(NewEventWithType(EventFriendList, time.Now()))

	// plan an update to the player's ignored list
	ignored := response.NewIgnoredListResponse(pe.player.Ignored)
	pe.PlanEvent(NewSendResponseEvent(ignored, time.Now()))

	// plan a welcome message
	msg := response.NewServerMessageResponse(g.welcomeMessage)
	pe.PlanEvent(NewSendResponseEvent(msg, time.Now()))

	// plan the first continuous idle check event
	pe.PlanEvent(NewEventWithType(EventCheckIdle, time.Now().Add(playerMaxIdleInterval)))
}

// RemovePlayer removes a previously joined player from the world.
func (g *Game) RemovePlayer(p *model.Player) {
	pe, unlockFunc := g.findPlayerAndLockGame(p)
	defer unlockFunc()

	if pe == nil {
		return
	}

	// add this player to the removal list, and let the next state update actually remove them
	g.removePlayers = append(g.removePlayers, pe)
}

// DoTakeGroundItem handles a player's request to pick up a ground item at a position, in global coordinates.
func (g *Game) DoTakeGroundItem(p *model.Player, itemID int, globalPos model.Vector2D) {
	var targetItem *model.Item
	for _, item := range g.items {
		if item.ID == itemID {
			targetItem = item
			break
		}
	}

	// validate the item is known
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
	var targetItem *model.Item
	for _, item := range g.items {
		if item.ID == itemID {
			targetItem = item
			break
		}
	}

	// validate the item is known
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

// broadcastPlayerStatus sends updates to other players that have them on their friends lists. An optional list of
// target player usernames can be passed to limit who receives the update.
//
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

// disconnect tells a player's client to disconnect from the server and terminates their connection.
// Concurrency requirements: none (any locks may be held).
func (g *Game) disconnect(pe *playerEntity) error {
	err := pe.writer.WriteUint8(response.DisconnectResponseHeader)
	pe.doneChan <- true

	return err
}

// loop continuously runs the main game server update cycle.
func (g *Game) loop() {
	for {
		select {
		case <-g.doneChan:
			logger.Infof("stopping game engine")
			return
		case <-g.ticker.C:
			err := g.handleGameUpdate()
			if err != nil {
				logger.Errorf("ending game state update due to error: %s", err)
				return
			}
		}
	}
}

// playerLoop continuously runs a game update cycle for a single player.
// Concurrency requirements: (a) game state should NOT be locked and (b) this player should NOT be locked.
func (g *Game) playerLoop(pe *playerEntity) {
	for {
		select {
		case <-pe.doneChan:
			// terminate this player's loop
			return

		case <-pe.changeChan:
			// a new event was planned; rerun the loop and let the scheduler report the next process time

		case update := <-pe.updateChan:
			// send a game state update, which takes priority over other pending events
			// TODO: should this be an event itself instead?
			err := update.Write(pe.writer)
			if err != nil {
				logger.Errorf("ending player loop due to error on update: %s", err)
				return
			}

		case <-time.After(pe.scheduler.TimeUntil()):
			// handle an event that is now ready for processing
			err := g.handlePlayerEvent(pe)
			if err != nil {
				logger.Errorf("ending player loop due to error: %s", err)
				return
			}
		}
	}
}

// loadAssets reads and parses all game asset.
// Concurrency requirements: none (any locks may be held).
func (g *Game) loadAssets(assetDir string) error {
	var err error
	manager := asset.NewManager(assetDir)

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
	g.items, err = manager.Items()
	if err != nil {
		return err
	}

	// FIXME: spawn some ground items for testing
	g.worldMap.Tile(model.Vector3D{X: 3209, Y: 3433}).AddItem(2)
	g.worldMap.Tile(model.Vector3D{X: 3209, Y: 3432}).AddItem(1)

	//g.worldMap.Tile(model.Vector3D{X: 3209, Y: 3429}).AddItem(54)

	//g.worldMap.Tile(model.Vector3D{X: 3213, Y: 3423}).AddItem(54)
	//g.worldMap.Tile(model.Vector3D{X: 3213, Y: 3424}).AddItem(249)

	//g.worldMap.Tile(model.Vector3D{X: 3242, Y: 3429}).AddItem(1052)
	//g.worldMap.Tile(model.Vector3D{X: 3243, Y: 3429}).AddItem(1187)
	//g.worldMap.Tile(model.Vector3D{X: 3243, Y: 3430}).AddItem(775)
	//g.worldMap.Tile(model.Vector3D{X: 3242, Y: 3430}).AddItem(861)
	//g.worldMap.Tile(model.Vector3D{X: 3241, Y: 3430}).AddItem(560)
	//g.worldMap.Tile(model.Vector3D{X: 3241, Y: 3429}).AddItem(962)
	//g.worldMap.Tile(model.Vector3D{X: 3241, Y: 3428}).AddItem(1053)
	//g.worldMap.Tile(model.Vector3D{X: 3242, Y: 3428}).AddItem(2550)

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

// handleChatCommand processes a chat command sent by a player.
// Concurrency requirements: (a) game state should NOT be locked and (b) this player should NOT be locked.
func (g *Game) handleChatCommand(pe *playerEntity, command *ChatCommand) {
	switch command.Type {
	case ChatCommandTypeSpawnItem:
		// prevent invalid items from being spawned
		if command.SpawnItem.ItemID >= 0 && command.SpawnItem.ItemID <= len(g.items) {
			g.mapManager.AddGroundItem(command.SpawnItem.ItemID, command.SpawnItem.DespawnTimeSeconds, pe.player.GlobalPos)
		} else {
			pe.PlanEvent(NewSendResponseEvent(
				response.NewServerMessageResponse(fmt.Sprintf("Invalid item: %d", command.SpawnItem.ItemID)),
				time.Now()))
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
		pe.PlanEvent(NewSendResponseEvent(response.NewServerMessageResponse(msg), time.Now()))
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
func (g *Game) addPlayerInventoryItem(pe *playerEntity, item *model.Item) {
	if item.Stackable {
		// TODO
		return
	}

	// find the next available slot in the player's inventory, if one exists
	slot := pe.player.NextFreeInventorySlot()
	if slot == -1 {
		pe.PlanEvent(NewSendResponseEvent(response.NewServerMessageResponse("You cannot carry any more items"), time.Now()))
		return
	}

	// set the item on the slot
	pe.player.SetInventoryItem(item, 1, slot)

	// update the player's inventory
	// FIXME: the interface id should not be hardcoded
	inventory := response.NewSetInventoryItemResponse(3214)
	inventory.AddSlot(slot, item.ID, 1)
	pe.PlanEvent(NewSendResponseEvent(inventory, time.Now()))
}

// removePlayerInventoryItem removes the first occurrence of an item from the player's inventory, and adds it to the
// world map.
// Concurrency requirements: (a) game state may be locked and (b) this player should be locked.
func (g *Game) removePlayerInventoryItem(pe *playerEntity, item *model.Item) *model.InventorySlot {
	slot := pe.player.InventorySlotWithItem(item.ID)
	if slot == nil {
		return nil
	}

	// remove the item from the player's inventory
	pe.player.ClearInventoryItem(slot.ID)

	// update the player's inventory
	inventory := response.NewSetInventoryItemResponse(3214)
	inventory.AddSlot(slot.ID, -1, 0)
	pe.PlanEvent(NewSendResponseEvent(inventory, time.Now()))

	return slot
}

// handleGameUpdate performs a game state update.
// Concurrency requirements: (a) game state should NOT be locked and (b) all players should NOT be locked.
func (g *Game) handleGameUpdate() error {
	g.mu.Lock()

	// determine if it's time to send game state updates to players
	sendUpdates := time.Now().Sub(g.lastPlayerUpdate) >= playerUpdateInterval

	// lock all players
	for _, pe := range g.players {
		pe.mu.Lock()
	}

	// reconcile the map state
	mapUpdates := g.mapManager.Reconcile()

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

			// mark the player as offline and broadcast their status
			g.playersOnline.Delete(pe.player.Username)
			g.broadcastPlayerStatus(pe)

			// flag the player's event loop to stop
			pe.doneChan <- true
		}
	}

	g.removePlayers = nil

	// update each player's own movements and send related updates
	for _, pe := range g.players {
		// prepare a new player update or use the pending, existing one
		if pe.nextUpdate == nil {
			pe.nextUpdate = response.NewPlayerUpdateResponse(pe.player.ID)
		}
		update := pe.nextUpdate

		// track if this player has moved to a new region
		hasChangedRegions := false

		// has this player teleported to a new location?
		if pe.teleportGlobal != nil {
			pe.player.GlobalPos = *pe.teleportGlobal
			origin, relative := g.playerRegionPosition(pe)

			if origin != pe.regionOrigin {
				pe.regionOrigin = origin

				region := response.NewLoadRegionResponse(pe.regionOrigin)
				pe.PlanEvent(NewSendResponseEvent(region, time.Now()))
			}

			relocate := response.NewPlayerUpdateResponse(pe.player.ID)
			relocate.SetLocalPlayerPosition(relative, true)
			pe.PlanEvent(NewSendResponseEvent(relocate, time.Now()))

			hasChangedRegions = true
			pe.teleportGlobal = nil
		}

		// check if the player is walking, and it's time to move to the next waypoint
		if pe.Walking() && time.Now().Sub(pe.lastWalkTime) >= playerWalkInterval {
			next := pe.path[pe.nextPathIdx]

			// add the change in direction to the local player's movement
			dir := model.DirectionFromDelta(next.Sub(pe.player.GlobalPos.To2D()))
			update.SetLocalPlayerWalk(dir)

			// update the player's position
			pe.player.GlobalPos.X = next.X
			pe.player.GlobalPos.Y = next.Y

			// move past this path segment and mark this as the last time the player was moved
			pe.nextPathIdx++
			pe.lastWalkTime = time.Now()

			// check if the player has moved into a new map region, and schedule a map region load is that's the case
			origin := g.findEffectiveRegion(pe)
			if origin != pe.regionOrigin {
				region := response.NewLoadRegionResponse(origin)
				pe.PlanEvent(NewSendResponseEvent(region, time.Now()))

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
					pe.PlanEvent(NewSendMultipleResponsesEvent(state, time.Now()))
				}

				// mark this as the current region the player's client has loaded
				pe.regionOrigin = origin
				hasChangedRegions = true
			}
		}

		// handle a deferred action for the player
		if pe.deferredAction != nil {
			switch pe.deferredAction.actionType {
			case pendingActionTakeGroundItem:
				action := pe.deferredAction.takeGroundItem

				// pick up a ground item only if the player has reached the position of that item
				if pe.player.GlobalPos != action.globalPos {
					break
				}

				// remove the ground item if it still exists, and allow the next reconciliation to take care of
				// updating the state of the map
				if g.mapManager.RemoveGroundItem(action.item.ID, action.globalPos) {
					// add the item to the player's inventory
					g.addPlayerInventoryItem(pe, action.item)
				}

				pe.deferredAction = nil

			case pendingActionDropInventoryItem:
				action := pe.deferredAction.dropInventoryItemAction

				// remove the item from the player's inventory
				slot := g.removePlayerInventoryItem(pe, action.item)
				if slot != nil {
					// put the item on the tile the player's standing on, and let the next reconciliation take care of
					// updating the state of the map
					timeout := int(itemDespawnInterval.Seconds())
					g.mapManager.AddGroundItem(slot.Item.ID, &timeout, pe.player.GlobalPos)
				}

				pe.deferredAction = nil
			}
		}

		// broadcast this player's status to friends and other target players if required
		if pe.nextStatusBroadcast != nil {
			g.broadcastPlayerStatus(pe, pe.nextStatusBroadcast.targets...)
			pe.nextStatusBroadcast = nil
		}

		// is this player in a region that has map updates? only send updates if they have not left this region
		regionGlobal := util.RegionOriginToGlobal(g.findEffectiveRegion(pe))
		if updates, ok := mapUpdates[regionGlobal]; ok && !hasChangedRegions {
			pe.PlanEvent(NewSendMultipleResponsesEvent(updates, time.Now()))
		}
	}

	// update each player with nearby players' updates
	for _, pe := range g.players {
		update := pe.nextUpdate

		// find players within visual distance of this player
		others := g.findSpectators(pe)
		for _, other := range others {
			// is this player new to us? if so we need to send an initial position and appearance update
			if _, ok := pe.tracking[other.player.ID]; !ok {
				posOffset := other.player.GlobalPos.Sub(pe.player.GlobalPos).To2D()
				update.AddToPlayerList(other.player.ID, posOffset, true, true)

				update.AddAppearanceUpdate(other.player.ID, other.player.Username, other.player.Appearance)
				pe.tracking[other.player.ID] = other
			} else if !update.Tracking(other.player.ID) {
				theirUpdate := other.nextUpdate

				// if the other player does not have an update, do not change their posture relative to us. otherwise
				// synchronize with their local movement
				if theirUpdate == nil {
					update.AddOtherPlayerNoUpdate(other.player.ID)
				} else {
					update.SyncLocalMovement(other.player.ID, theirUpdate)
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

		// unlock the player and send an update if needed
		pe.mu.Unlock()
		if sendUpdates {
			pe.updateChan <- pe.nextUpdate
			pe.nextUpdate = nil
		}
	}

	if sendUpdates {
		g.lastPlayerUpdate = time.Now()
	}

	g.mu.Unlock()
	return nil
}

// handlePlayerEvent processes the next scheduled event for a player.
// Concurrency requirements: (a) game state should NOT be locked and (b) this player should NOT be locked.
func (g *Game) handlePlayerEvent(pe *playerEntity) error {
	// get the next scheduled event, if any
	event := pe.scheduler.Next()
	if event == nil {
		return nil
	}

	switch event.Type {
	case EventCheckIdle, EventCheckIdleImmediate:
		// determine if the player has been idle for too long, and if so disconnect them
		if time.Now().Sub(pe.lastInteraction) >= playerMaxIdleInterval {
			_ = g.disconnect(pe)
			return nil
		}

		// schedule the next idle check event if this check was not on-demand
		if event.Type != EventCheckIdleImmediate {
			pe.scheduler.Plan(NewEventWithType(EventCheckIdle, time.Now().Add(playerMaxIdleInterval)))
		}

	case EventFriendList:
		// send a player their entire friends list

		// tell the client the list is loading
		status := response.NewFriendsListStatusResponse(model.FriendsListStatusLoading)
		err := status.Write(pe.writer)
		if err != nil {
			return err
		}

		// send the status for each friends list entry
		for _, username := range pe.player.Friends {
			var friend *response.FriendStatusResponse

			// send a status for this player if they are online or not
			_, ok := g.playersOnline.Load(username)
			if ok {
				friend = response.NewFriendStatusResponse(username, 69)
			} else {
				friend = response.NewOfflineFriendStatusResponse(username)
			}

			err := friend.Write(pe.writer)
			if err != nil {
				return err
			}
		}

		// finally, tell the client the list has been sent
		status = response.NewFriendsListStatusResponse(model.FriendsListStatusLoaded)
		err = status.Write(pe.writer)
		if err != nil {
			return err
		}

	case EventUpdateTabInterfaces:
		// send all client tab interface ids
		for _, tab := range model.ClientTabs {
			// does the player have an interface for this tab?
			var r *response.SidebarInterfaceResponse
			if id, ok := pe.tabInterfaces[tab]; ok {
				r = response.NewSidebarInterfaceResponse(tab, id)
			} else {
				r = response.NewRemoveSidebarInterfaceResponse(tab)
			}

			err := r.Write(pe.writer)
			if err != nil {
				return err
			}
		}

	case EventSkills:
		// send each skill to the client
		for _, skill := range pe.player.Skills {
			resp := response.NewSkillDataResponse(skill)
			err := resp.Write(pe.writer)
			if err != nil {
				return err
			}
		}

	case EventSendResponse:
		// send a generic response to the client
		err := event.Responses[0].Write(pe.writer)
		if err != nil {
			return err
		}

	case EventSendManyResponses:
		// send all responses to the client
		for _, resp := range event.Responses {
			err := resp.Write(pe.writer)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// sendPlayerUpdate sends a game state update to the player.
// Concurrency requirements: (a) game state may be locked (b) this player should NOT be locked.
func (g *Game) sendPlayerUpdate(pe *playerEntity) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if pe.nextUpdate != nil {
		err := pe.nextUpdate.Write(pe.writer)
		if err != nil {
			return err
		}

		pe.nextUpdate = nil
	}

	return nil
}
