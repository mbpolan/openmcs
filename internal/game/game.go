package game

import (
	"github.com/mbpolan/openmcs/internal/asset"
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

// Game is the game engine and representation of the game world.
type Game struct {
	items         []*model.Item
	doneChan      chan bool
	ticker        *time.Ticker
	objects       []*model.WorldObject
	players       []*playerEntity
	playersOnline sync.Map
	worldMap      *model.Map
	mu            sync.RWMutex
}

// NewGame creates a new game engine using game assets located at the given assetDir.
func NewGame(assetDir string) (*Game, error) {
	g := &Game{
		doneChan: make(chan bool, 1),
	}

	// load game asset
	err := g.loadAssets(assetDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load game asset")
	}

	return g, nil
}

func (g *Game) Stop() {
	g.ticker.Stop()
	g.doneChan <- true
}

func (g *Game) Run() {
	g.ticker = time.NewTicker(50 * time.Millisecond)
	go g.loop()
}

// AddFriend attempts to add another player to the player's friends list.
func (g *Game) AddFriend(p *model.Player, username string) {
	target := strings.Trim(strings.ToLower(username), " ")

	// TODO: validate if target player exists in persistent storage and get their properly cased name

	// TODO: update friends list in persistent storage

	// we need to manually control the player's lock
	pe, unlockFunc := g.findPlayerAndLockGame(p)
	defer unlockFunc()
	if pe == nil {
		return
	}

	pe.mu.Lock()

	// is this player already in their friend's list
	exists := false
	for _, u := range pe.player.Friends {
		if strings.ToLower(u) == target {
			exists = true
		}
	}

	// avoid adding duplicates
	if exists {
		pe.mu.Unlock()
		return
	}

	pe.player.Friends = append(pe.player.Friends, target)
	pe.mu.Unlock()

	// send a status update about the player that was just added
	var tpe *playerEntity
	for _, other := range g.players {
		if other.player.Username == target {
			tpe = other
			break
		}
	}

	if tpe != nil {
		g.broadcastPlayerStatus(tpe, pe.player.Username)
	}
}

// RemoveFriend removes another player from the player's friends list.
func (g *Game) RemoveFriend(p *model.Player, username string) {
	// only lock the game state temporarily
	g.mu.Lock()
	pe := g.findPlayer(p)
	g.mu.Unlock()

	if pe == nil {
		return
	}

	pe.mu.Lock()

	// remove the player from the friends list
	target := strings.Trim(strings.ToLower(username), " ")
	for i, other := range pe.player.Friends {
		if strings.ToLower(other) == target {
			pe.player.Friends = append(pe.player.Friends[:i], pe.player.Friends[i+1:]...)
			break
		}
	}

	// the client automatically removes players so we don't need to send an explicit update
	pe.mu.Unlock()
}

// SetPlayerModes updates the chat and interaction modes for a player.
func (g *Game) SetPlayerModes(p *model.Player, publicChat model.ChatMode, privateChat model.ChatMode, interaction model.InteractionMode) {
	pe, unlockFunc := g.findPlayerAndLockAll(p)
	defer unlockFunc()
	if pe == nil {
		return
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

// RequestLogout attempts to log out a player.
func (g *Game) RequestLogout(p *model.Player, action int) {
	pe, unlockFunc := g.findPlayerAndLockAll(p)
	defer unlockFunc()
	if pe == nil {
		return
	}

	// TODO: check if player can be logged out (ie: are they in combat, etc.)
	_ = g.disconnect(pe)
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

// WalkPlayer starts moving the player to a destination from a start position then following a set of waypoints. The
// slice of waypoints are deltas relative to start.
func (g *Game) WalkPlayer(p *model.Player, start model.Vector2D, waypoints []model.Vector2D) {
	pe, unlockFunc := g.findPlayerAndLockAll(p)
	defer unlockFunc()
	if pe == nil {
		return
	}

	// the starting position is offset by 6 tiles from the region origin, and serves as the basis for waypoints
	initial := model.Vector2D{
		X: start.X + (util.MapScale3D.X * 6),
		Y: start.Y + (util.MapScale3D.Y * 6),
	}

	// convert each waypoint into global coordinates
	actuals := []model.Vector2D{initial}
	for _, w := range waypoints {
		actuals = append(actuals, model.Vector2D{
			X: initial.X + w.X,
			Y: initial.Y + w.Y,
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

// AddPlayer joins a player to the world and handles ongoing game events and network interactions.
func (g *Game) AddPlayer(p *model.Player, writer *network.ProtocolWriter) {
	pe := newPlayerEntity(p, writer)

	// set initial client tab interfaces
	// TODO: these ids should not be hardcoded
	pe.tabInterfaces = map[model.ClientTab]int{
		model.ClientTabFriendsList: 5065,
		model.ClientTabLogout:      2449,
	}

	g.mu.Lock()

	// add the player to the player list
	g.players = append(g.players, pe)

	// mark the player as being online and broadcast their status. the game state needs to be locked up front.
	g.playersOnline.Store(pe.player.Username, true)
	g.broadcastPlayerStatus(pe)

	g.mu.Unlock()

	// start the player's event loop
	go g.playerLoop(pe)

	// plan an initial map region load
	region := response.NewLoadRegionResponse(util.GlobalToRegionOrigin(p.GlobalPos).To2D())
	pe.PlanEvent(NewSendResponseEvent(region, time.Now()))

	// plan an initial player update
	update := response.NewPlayerUpdateResponse(p.ID)
	update.SetLocalPlayerPosition(util.GlobalToRegionLocal(p.GlobalPos), true)
	update.AddAppearanceUpdate(p.ID, p.Username, p.Appearance)
	pe.PlanEvent(NewSendResponseEvent(update, time.Now()))

	// plan an update to the client sidebar interfaces
	pe.PlanEvent(NewEventWithType(EventUpdateTabInterfaces, time.Now()))

	// plan an update to the client's interaction modes
	modes := response.NewSetModesResponse(pe.player.Modes.PublicChat, pe.player.Modes.PrivateChat, pe.player.Modes.Interaction)
	pe.PlanEvent(NewSendResponseEvent(modes, time.Now()))

	// plan an update to the player's friends list
	pe.PlanEvent(NewEventWithType(EventFriendList, time.Now()))

	// plan the first continuous player update after the initial one is done
	pe.PlanEvent(NewEventWithType(EventPlayerUpdate, time.Now().Add(playerUpdateInterval)))

	// plan the first continuous idle check event
	pe.PlanEvent(NewEventWithType(EventCheckIdle, time.Now().Add(playerMaxIdleInterval)))
}

// RemovePlayer removes a previously joined player from the world.
func (g *Game) RemovePlayer(p *model.Player) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// remove the player from the game
	for i, pe := range g.players {
		if pe.player.ID == p.ID {
			// drop the player from the player list
			g.players = append(g.players[:i], g.players[i+1:]...)

			// mark the player as offline and broadcast their status
			g.playersOnline.Delete(pe.player.Username)
			g.broadcastPlayerStatus(pe)

			// flag the player's event loop to stop
			pe.doneChan <- true
		} else {
			delete(pe.tracking, p.ID)
		}
	}
}

// broadcastPlayerStatus sends updates to other players that have them on their friends lists. An optional list of
// target player usernames can be passed to limit who receives the update. The game state should be locked before
// calling this method.
func (g *Game) broadcastPlayerStatus(pe *playerEntity, targets ...string) {
	_, online := g.playersOnline.Load(pe.player.Username)

	// find players that have this player on their friends list
	for _, other := range g.players {
		// skip the same player, or if they are not targeted
		if pe == other || (len(targets) > 0 && !util.Contains(targets, other.player.Username)) {
			continue
		}

		other.mu.Lock()

		if util.Contains(other.player.Friends, pe.player.Username) {
			var update *response.FriendStatusResponse
			if online {
				update = response.NewFriendStatusResponse(pe.player.Username, 69)
			} else {
				update = response.NewOfflineFriendStatusResponse(pe.player.Username)
			}

			err := update.Write(other.writer)
			if err != nil {
				// other player could have disconnected, so we don't treat this as a fatal error
				logger.Debugf("failed to send friend status update for %s to %s: %s", pe.player.Username, other.player.Username, err)
			}
		}

		other.mu.Unlock()
	}
}

// findSpectators returns a slice of players that are within visual distance of a given player.
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
func (g *Game) playerLoop(pe *playerEntity) {
	for {
		select {
		case <-pe.doneChan:
			// terminate this player's loop
			return

		case <-pe.resetChan:
			// a new event was planned; rerun the loop and let the scheduler report the next process time

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
func (g *Game) loadAssets(assetDir string) error {
	var err error
	manager := asset.NewManager(assetDir)

	// load map data
	g.worldMap, err = manager.Map()
	if err != nil {
		return err
	}

	// load world objects
	g.objects, err = manager.WorldObjects()
	if err != nil {
		return err
	}

	// load items
	g.items, err = manager.Items()
	if err != nil {
		return err
	}

	return nil
}

// handleGameUpdate performs a game state update.
func (g *Game) handleGameUpdate() error {
	g.mu.Lock()

	// lock all players
	for _, pe := range g.players {
		pe.mu.Lock()
	}

	// update each player's own movements
	for _, pe := range g.players {
		// prepare a new player update or use the pending, existing one
		if pe.nextUpdate == nil {
			pe.nextUpdate = response.NewPlayerUpdateResponse(pe.player.ID)
		}
		update := pe.nextUpdate

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
		}

		pe.nextUpdate = update
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
				receive := true

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

	// unlock all players
	for _, pe := range g.players {
		pe.mu.Unlock()
	}

	g.mu.Unlock()
	return nil
}

// handlePlayerEvent processes the next scheduled event for a player.
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

	case EventPlayerUpdate:
		// send a player update
		err := g.sendPlayerUpdate(pe)
		if err != nil {
			return err
		}

		// plan the next update
		pe.scheduler.Plan(&Event{
			Type:     EventPlayerUpdate,
			Schedule: time.Now().Add(playerUpdateInterval),
		})

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

	case EventSendResponse:
		// send a generic response to the client
		err := event.Response.Write(pe.writer)
		if err != nil {
			return err
		}
	}

	return nil
}

// sendPlayerUpdate sends a game state update to the player.
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
