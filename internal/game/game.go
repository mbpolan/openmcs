package game

import (
	"github.com/mbpolan/openmcs/internal/asset"
	"github.com/mbpolan/openmcs/internal/asset/loader"
	"github.com/mbpolan/openmcs/internal/logger"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/pkg/errors"
	"time"
)

// Game is the game engine and representation of the game world.
type Game struct {
	items    []*model.Item
	doneChan chan bool
	ticker   *time.Ticker
	objects  []*model.WorldObject
}

// NewGame creates a new game engine.
func NewGame() (*Game, error) {
	g := &Game{
		doneChan: make(chan bool, 1),
	}

	// load game asset
	err := g.loadAssets()
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

func (g *Game) loop() {
	for {
		select {
		case <-g.doneChan:
			logger.Infof("stopping game engine")
			return
		case <-g.ticker.C:
		}
	}
}

// loadAssets reads and parses all game asset.
func (g *Game) loadAssets() error {
	c := asset.NewCacheFile("./data", 0)
	a, err := c.Archive(2)
	if err != nil {
		return err
	}

	// load world objects
	woLoader := loader.NewWorldObjectLoader(a)
	g.objects, err = woLoader.Load()
	if err != nil {
		return err
	}

	// load items
	itemLoader := loader.NewItemLoader(a)
	g.items, err = itemLoader.Load()
	if err != nil {
		return err
	}

	return nil
}
