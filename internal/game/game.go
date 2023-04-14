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
const playerUpdateInterval = 600 * time.Millisecond

type playerEntity struct {
	lastInteraction time.Time
	player          *model.Player
	resetChan       chan bool
	doneChan        chan bool
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
	for _, pe := range g.players {
		if pe.player.ID == p.ID {
			pe.lastInteraction = time.Now()
			break
		}
	}
}

// MarkPlayerInactive flags that a player's client reported them as being idle.
func (g *Game) MarkPlayerInactive(p *model.Player) {
	for _, pe := range g.players {
		if pe.player.ID == p.ID {
			pe.scheduler.Plan(NewEventWithType(EventCheckIdleImmediate, time.Now()))
			break
		}
	}
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
	go g.playerLoop(pe)

	// plan an initial map region load
	region := response.NewLoadRegionResponse(util.GlobalToRegionOrigin(p.GlobalPos).To2D())
	pe.PlanEvent(NewSendResponseEvent(region, time.Now()))

	// plan an initial player update
	update := response.NewPlayerUpdateResponse()
	update.SetLocalPlayerPosition(util.GlobalToRegionLocal(p.GlobalPos), true, true)
	update.AddAppearanceUpdate(p.ID, p.Username, p.Appearance)
	pe.PlanEvent(NewSendResponseEvent(update, time.Now()))

	// plan the first continuous player update after the initial one is done
	pe.PlanEvent(NewEventWithType(EventPlayerUpdate, time.Now().Add(playerUpdateInterval)))

	// plan the first continuous idle check event
	pe.PlanEvent(NewEventWithType(EventCheckIdle, time.Now().Add(playerMaxIdleInterval)))

	g.players = append(g.players, pe)
}

// RemovePlayer removes a previously joined player from the world.
func (g *Game) RemovePlayer(p *model.Player) {
	for i, pe := range g.players {
		if pe.player == p {
			pe.doneChan <- true
			g.players = append(g.players[:i], g.players[i+1:]...)
			break
		}
	}
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
			_ = pe.writer.WriteByte(response.DisconnectResponseHeader)
			pe.doneChan <- true

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

// sendPlayerUpdate sends a game state update to the player.
func (g *Game) sendPlayerUpdate(pe *playerEntity) error {
	resp := response.NewPlayerUpdateResponse()

	// TODO: update player's actual current state
	err := resp.Write(pe.writer)
	if err != nil {
		return err
	}

	return nil
}
