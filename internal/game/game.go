package game

import (
	"github.com/mbpolan/openmcs/internal/asset"
	"github.com/mbpolan/openmcs/internal/logger"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/network/response"
	"github.com/mbpolan/openmcs/internal/util"
	"github.com/pkg/errors"
	"time"
)

// playerMaxIdleInterval is the maximum time a player can be idle before being forcefully logged out.
// TODO: this should be configurable
const playerMaxIdleInterval = 3 * time.Minute

// playerUpdateInterval defines how often player updates are sent.
const playerUpdateInterval = 200 * time.Millisecond

type playerEntity struct {
	lastInteraction time.Time
	player          *model.Player
	resetChan       chan bool
	doneChan        chan bool
	path            []model.Vector2D
	scheduler       *Scheduler
	writer          *network.ProtocolWriter
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
	pe := g.findPlayerByID(p)
	if pe == nil {
		return
	}

	pe.lastInteraction = time.Now()
}

// MarkPlayerInactive flags that a player's client reported them as being idle.
func (g *Game) MarkPlayerInactive(p *model.Player) {
	pe := g.findPlayerByID(p)
	if pe == nil {
		return
	}

	pe.scheduler.Plan(NewEventWithType(EventCheckIdleImmediate, time.Now()))
}

func (g *Game) RequestLogout(p *model.Player, action int) {
	pe := g.findPlayerByID(p)
	if pe == nil {
		return
	}

	// TODO: check if player can be logged out (ie: are they in combat, etc.)
	_ = g.disconnect(pe)
}

// WalkPlayer starts moving the player to a specific destination via waypoints.
func (g *Game) WalkPlayer(p *model.Player, waypoints []model.Vector2D) {
	if len(waypoints) == 0 {
		return
	}

	pe := g.findPlayerByID(p)
	if pe == nil {
		return
	}

	// the first waypoint coordinate is offset by 6 tiles from the region origin
	waypoints[0].X = (waypoints[0].X + (util.MapScale3D.X)*6) - p.GlobalPos.X
	waypoints[0].Y = (waypoints[0].Y + (util.MapScale3D.Y)*6) - p.GlobalPos.Y

	// start traversing from the player's current position
	from := p.GlobalPos.To2D()
	var path []model.Vector2D

	// plan a direct path to each waypoint
	for _, w := range waypoints {
		dx := w.X
		dy := w.Y

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
	g.players = append(g.players, pe)
	go g.playerLoop(pe)

	// plan an initial map region load
	region := response.NewLoadRegionResponse(util.GlobalToRegionOrigin(p.GlobalPos).To2D())
	pe.PlanEvent(NewSendResponseEvent(region, time.Now()))

	// plan an initial player update
	update := response.NewPlayerUpdateResponse()
	update.SetLocalPlayerPosition(util.GlobalToRegionLocal(p.GlobalPos), true, true)
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
	for i, pe := range g.players {
		if pe.player == p {
			g.players = append(g.players[:i], g.players[i+1:]...)
			pe.doneChan <- true
			break
		}
	}
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
			// TODO: update game state
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

		default:
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
	update := response.NewPlayerUpdateResponse()

	// update player's complete current state
	// TODO: handle remaining possibilities
	if len(pe.path) != 0 {
		// the player is walking
		next := pe.path[0]

		// find the next direction to walk towards
		dir := model.DirectionFromDelta(next.Sub(pe.player.GlobalPos.To2D()))
		if dir != model.DirectionNone {
			update.SetLocalPlayerWalk(dir, false)
		}

		// remove this waypoint from the path and update the player's position
		pe.path = pe.path[1:]
		pe.player.GlobalPos.X = next.X
		pe.player.GlobalPos.Y = next.Y
	}

	err := update.Write(pe.writer)
	if err != nil {
		return err
	}

	return nil
}
