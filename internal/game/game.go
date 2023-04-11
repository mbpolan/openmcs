package game

import (
	"github.com/mbpolan/openmcs/internal/assets"
	"github.com/mbpolan/openmcs/internal/assets/loaders"
	"github.com/mbpolan/openmcs/internal/logger"
	"github.com/mbpolan/openmcs/internal/models"
	"github.com/pkg/errors"
	"time"
)

// Game is the game engine and representation of the game world.
type Game struct {
	doneChan chan bool
	ticker   *time.Ticker
	objects  []*models.WorldObject
}

// NewGame creates a new game engine.
func NewGame() (*Game, error) {
	g := &Game{
		doneChan: make(chan bool, 1),
	}

	// load game assets
	err := g.loadAssets()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load game assets")
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

// loadAssets reads and parses all game assets.
func (g *Game) loadAssets() error {
	c := assets.NewCacheFile("./data", 0)
	a, err := c.Archive(2)
	if err != nil {
		return err
	}

	// load world objects
	woLoader := loaders.NewWorldObjectLoader(a)
	g.objects, err = woLoader.Load()
	if err != nil {
		return err
	}

	return nil
}
