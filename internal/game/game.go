package game

import (
	"github.com/mbpolan/openmcs/internal/logger"
	"time"
)

// Game is the game engine and representation of the game world.
type Game struct {
	ticker   *time.Ticker
	doneChan chan bool
}

// NewGame creates a new game engine.
func NewGame() *Game {
	return &Game{
		doneChan: make(chan bool, 1),
	}
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
