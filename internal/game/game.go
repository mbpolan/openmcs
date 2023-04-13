package game

import (
	"github.com/mbpolan/openmcs/internal/asset"
	"github.com/mbpolan/openmcs/internal/logger"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/network/responses"
	"github.com/mbpolan/openmcs/internal/util"
	"github.com/pkg/errors"
	"time"
)

type playerEntity struct {
	player   *model.Player
	doneChan chan bool
	ticker   *time.Ticker
	writer   *network.ProtocolWriter
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

func (g *Game) AddPlayer(p *model.Player, writer *network.ProtocolWriter) {
	pe := &playerEntity{
		player:   p,
		doneChan: make(chan bool, 1),
		ticker:   time.NewTicker(200 * time.Millisecond),
		writer:   writer,
	}

	// send an initial map load
	// TODO: schedule this instead
	region := responses.NewLoadRegionResponse(util.GlobalToRegionOrigin(p.GlobalPos).To2D())
	_ = region.Write(writer)

	// send an initial player update
	// TODO: schedule this instead
	update := responses.NewPlayerUpdateResponse()
	update.SetLocalPlayerPosition(util.GlobalToRegionLocal(p.GlobalPos), true, true)
	_ = update.Write(writer)

	g.players = append(g.players, pe)
	go g.playerLoop(pe)
}

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
			return
		case <-pe.ticker.C:
			err := g.sendPlayerUpdate(pe)
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

func (g *Game) sendPlayerUpdate(pe *playerEntity) error {
	resp := responses.NewPlayerUpdateResponse()

	// TODO: update player's actual current state
	resp.SetLocalPlayerNoMovement()

	err := resp.Write(pe.writer)
	if err != nil {
		return err
	}

	return nil
}
