package game

import (
	"github.com/mbpolan/openmcs/internal/asset"
	"github.com/mbpolan/openmcs/internal/logger"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/network/response"
	"github.com/mbpolan/openmcs/internal/util"
	"github.com/pkg/errors"
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

type trackedPlayerEntity struct {
	pe                  *playerEntity
	needsUpdate         bool
	lastPosition        model.Vector3D
	lastTrackedWalkTime time.Time
	lastChatMessage     *model.ChatMessage
	lastTrackedChatTime time.Time
}

type playerEntity struct {
	lastInteraction   time.Time
	player            *model.Player
	tracking          []*trackedPlayerEntity
	resetChan         chan bool
	doneChan          chan bool
	path              []model.Vector2D
	scheduler         *Scheduler
	writer            *network.ProtocolWriter
	lastSentWalkTime  time.Time
	lastWalkTime      time.Time
	lastWalkDirection model.Direction
	lastChatMessage   *model.ChatMessage
	lastChatTime      time.Time
	isMidWalk         bool
}

// PlanEvent adds a scheduled event to this player's queue and resets the event timer.
func (pe *playerEntity) PlanEvent(e *Event) {
	pe.scheduler.Plan(e)
	pe.resetChan <- true
}

// Game is the game engine and representation of the game world.
type Game struct {
	items    []*model.Item
	doneChan chan bool
	ticker   *time.Ticker
	objects  []*model.WorldObject
	players  []*playerEntity
	worldMap *model.Map
	mu       sync.RWMutex
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

// MarkPlayerActive updates a player's last activity tracker and prevents them from becoming idle.
func (g *Game) MarkPlayerActive(p *model.Player) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	pe := g.findPlayerByID(p)
	if pe == nil {
		return
	}

	pe.lastInteraction = time.Now()
}

// MarkPlayerInactive flags that a player's client reported them as being idle.
func (g *Game) MarkPlayerInactive(p *model.Player) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	pe := g.findPlayerByID(p)
	if pe == nil {
		return
	}

	pe.scheduler.Plan(NewEventWithType(EventCheckIdleImmediate, time.Now()))
}

// RequestLogout attempts to log out a player.
func (g *Game) RequestLogout(p *model.Player, action int) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	pe := g.findPlayerByID(p)
	if pe == nil {
		return
	}

	// TODO: check if player can be logged out (ie: are they in combat, etc.)
	_ = g.disconnect(pe)
}

// DoPlayerChat broadcasts a player's chat message to nearby players.
func (g *Game) DoPlayerChat(p *model.Player, effect model.ChatEffect, color model.ChatColor, text string) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	pe := g.findPlayerByID(p)
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
	g.mu.RLock()
	defer g.mu.RUnlock()

	pe := g.findPlayerByID(p)
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
	pe.lastWalkTime = time.Now()
	pe.lastSentWalkTime = pe.lastWalkTime
}

// AddPlayer joins a player to the world and handles ongoing game events and network interactions.
func (g *Game) AddPlayer(p *model.Player, writer *network.ProtocolWriter) {
	pe := &playerEntity{
		lastInteraction: time.Now(),
		player:          p,
		resetChan:       make(chan bool),
		doneChan:        make(chan bool, 1),
		scheduler:       NewScheduler(),
		writer:          writer,
	}

	// start the player's processing loop
	g.mu.Lock()
	g.players = append(g.players, pe)
	g.mu.Unlock()

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
	g.planClientTabInterfaces(pe)

	// plan the first continuous player update after the initial one is done
	pe.PlanEvent(NewEventWithType(EventPlayerUpdate, time.Now().Add(playerUpdateInterval)))

	// plan the first continuous idle check event
	pe.PlanEvent(NewEventWithType(EventCheckIdle, time.Now().Add(playerMaxIdleInterval)))
}

// RemovePlayer removes a previously joined player from the world.
func (g *Game) RemovePlayer(p *model.Player) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// remove the player from the game. other players who were tracking this player will have their spectators list
	// refreshed the next time the game state update runs.
	for i, pe := range g.players {
		if pe.player == p {
			g.players = append(g.players[:i], g.players[i+1:]...)
			pe.doneChan <- true
			break
		}
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

// findPlayerByID returns the playerEntity for the corresponding player.
func (g *Game) findPlayerByID(p *model.Player) *playerEntity {
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

	// check each player in the world
	for _, pe := range g.players {
		// find players within visual distance of this player
		others := g.findSpectators(pe)

		// has this set of tracked players changed?
		trackingChanged := false
		if len(pe.tracking) == len(others) {
			for i, other := range others {
				if pe.tracking[i].pe.player.ID != other.player.ID {
					trackingChanged = true
					break
				}
			}
		} else {
			trackingChanged = true
		}

		// update this player's slice of tracked players
		if trackingChanged {
			// TODO: can we reduce the amount of allocations here?
			pe.tracking = make([]*trackedPlayerEntity, len(others))
			for i, other := range others {
				pe.tracking[i] = &trackedPlayerEntity{
					pe:           other,
					needsUpdate:  true,
					lastPosition: other.player.GlobalPos,
				}
			}
		}

		// check if the player is walking, and it's time to move to the next waypoint
		if len(pe.path) > 0 && time.Now().Sub(pe.lastWalkTime) >= playerWalkInterval/2 {
			if pe.isMidWalk {
				pe.isMidWalk = false
			} else {
				pe.isMidWalk = true

				next := pe.path[0]

				// find the direction the next segment is at relative to the player's current position
				pe.lastWalkDirection = model.DirectionFromDelta(next.Sub(pe.player.GlobalPos.To2D()))

				// update the player's position
				pe.player.GlobalPos.X = next.X
				pe.player.GlobalPos.Y = next.Y

				// drop this path segment and mark this as the last time the player was moved
				pe.path = pe.path[1:]
				pe.lastWalkTime = time.Now()
			}
		}

		logger.Debugf("%s has tracking: %+v", pe.player.Username, pe.tracking)
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

	case EventSendResponse:
		// send a generic response to the client
		err := event.Response.Write(pe.writer)
		if err != nil {
			return err
		}
	}

	return nil
}

// planClientTabInterfaces schedules an update for the player's client to refresh its tab interfaces.
func (g *Game) planClientTabInterfaces(pe *playerEntity) {
	// TODO: these ids should not be hardcoded
	r := response.NewSidebarInterfaceResponse(model.ClientTabLogout, 2449)
	pe.scheduler.Plan(NewSendResponseEvent(r, time.Now()))
}

// sendPlayerUpdate sends a game state update to the player.
func (g *Game) sendPlayerUpdate(pe *playerEntity) error {
	update := response.NewPlayerUpdateResponse(pe.player.ID)

	// update player's complete current state
	// TODO: handle remaining possibilities

	// update the player if they are walking, and it's time for their next walk increment
	if pe.lastWalkTime.After(pe.lastSentWalkTime) && pe.isMidWalk {
		update.SetLocalPlayerWalk(pe.lastWalkDirection)
		pe.lastSentWalkTime = pe.lastWalkTime
	}

	// update players that are in this player's visible range
	for _, other := range pe.tracking {
		// if the other player is walking, update their movement. otherwise just add them to the player list so they
		// are still tracked.
		if other.pe.lastWalkTime.After(other.lastTrackedWalkTime) {
			if other.pe.isMidWalk {
				update.AddOtherPlayerNoUpdate(other.pe.player.ID)
			} else {
				update.AddOtherPlayerWalk(other.pe.player.ID, other.pe.lastWalkDirection)
				other.lastTrackedWalkTime = other.pe.lastWalkTime
			}
		} else {
			// compute the other player's location relative to this player, and add them to the update list
			posOffset := other.pe.player.GlobalPos.Sub(pe.player.GlobalPos).To2D()
			update.AddToPlayerList(other.pe.player.ID, posOffset, true, true)
		}

		// if the other player needs additional updates, include those as well
		if other.needsUpdate {
			update.AddAppearanceUpdate(other.pe.player.ID, other.pe.player.Username, pe.player.Appearance)
			other.needsUpdate = false
		}

		// has the other player posted a new chat message?
		if other.pe.lastChatTime.After(other.lastTrackedChatTime) && other.pe.lastChatMessage != nil {
			update.AddChatMessage(other.pe.player.ID, other.pe.lastChatMessage)
			other.lastTrackedChatTime = other.pe.lastChatTime
		}
	}

	err := update.Write(pe.writer)
	if err != nil {
		return err
	}

	return nil
}
