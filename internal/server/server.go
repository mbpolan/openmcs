package server

import (
	"context"
	"fmt"
	"github.com/mbpolan/openmcs/internal/config"
	"github.com/mbpolan/openmcs/internal/game"
	"github.com/mbpolan/openmcs/internal/logger"
	"github.com/mbpolan/openmcs/internal/store"
	"github.com/mbpolan/openmcs/internal/util"
	"github.com/pkg/errors"
	"net"
	"sync"
)

// Server provides the network infrastructure for a game and login server.
type Server struct {
	config      *config.Config
	bindAddress string
	clients     []*ClientHandler
	closeChan   chan *ClientHandler
	store       *store.Store
	doneChan    chan bool
	listener    net.Listener
	game        *game.Game
	mu          sync.Mutex
	sessionKey  uint64
}

// New creates a server instance with a configuration.
func New(cfg *config.Config) (*Server, error) {
	return &Server{
		bindAddress: fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		clients:     nil,
		closeChan:   make(chan *ClientHandler),
		config:      cfg,
		doneChan:    make(chan bool, 1),
		mu:          sync.Mutex{},
		sessionKey:  0,
	}, nil
}

// Stop terminates the server and stops accepting new connections.
func (s *Server) Stop() {
	s.doneChan <- true
	s.listener.Close()
}

// Run begins listening for connections and spawning request handlers.
func (s *Server) Run() error {
	var err error

	// prepare the persistent store
	s.store, err = store.New(s.config)
	if err != nil {
		return errors.Wrap(err, "failed creating persistent store")
	}

	defer s.store.Close()

	// run migrations
	err = s.store.Migrate()
	if err != nil {
		return errors.Wrap(err, "failed running persistent store migrations")
	}

	// create a new game engine instance
	s.game, err = game.NewGame(s.config)
	if err != nil {
		return errors.Wrap(err, "failed creating game world")
	}

	// start listening for connections
	s.listener, err = net.Listen("tcp", s.bindAddress)
	if err != nil {
		return errors.Wrap(err, "failed to listen on socket")
	}

	// start the game engine loop
	s.game.Run()

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer s.listener.Close()
	defer s.game.Stop()
	defer cancelFunc()

	go s.cleanUpHandler(ctx)

	logger.Infof("server listening on %s", s.bindAddress)

	for {
		// listen for incoming connections, and gracefully exit if the listener has stopped
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.doneChan:
				return nil
			default:
				return errors.Wrap(err, "failed to accept connection")
			}
		}

		client := NewClientHandler(conn, s.closeChan, s.store, s.game, s.sessionKey)

		s.mu.Lock()
		s.clients = append(s.clients, client)
		s.mu.Unlock()

		go client.Handle()
	}
}

func (s *Server) cleanUpHandler(ctx context.Context) {
	for {
		select {
		case h := <-s.closeChan:
			logger.Infof("disconnecting player")

			s.mu.Lock()
			s.clients = util.Remove(s.clients, h)
			s.mu.Unlock()

		case <-ctx.Done():
			return
		}
	}
}
