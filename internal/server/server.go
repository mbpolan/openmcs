package server

import (
	"context"
	"fmt"
	"github.com/mbpolan/openmcs/internal/game"
	"github.com/mbpolan/openmcs/internal/logger"
	"github.com/mbpolan/openmcs/internal/util"
	"github.com/pkg/errors"
	"net"
	"sync"
)

// Server provides the network infrastructure for a game and login server.
type Server struct {
	assetDir    string
	bindAddress string
	clients     []*ClientHandler
	closeChan   chan *ClientHandler
	db          *game.Database
	doneChan    chan bool
	listener    net.Listener
	game        *game.Game
	mu          sync.Mutex
	sessionKey  uint64
}

// Options are configuration parameters that can be used to customize a server.
type Options struct {
	AssetDir string
	Address  string
	Port     int
}

// New creates a server instance with options..
func New(opts Options) (*Server, error) {
	return &Server{
		bindAddress: fmt.Sprintf("%s:%d", opts.Address, opts.Port),
		clients:     nil,
		closeChan:   make(chan *ClientHandler),
		db:          game.NewDatabase(),
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

// Run begins listening for connections and spawning requests handlers.
func (s *Server) Run() error {
	var err error

	// create a new game engine instance
	s.game, err = game.NewGame(s.assetDir)
	if err != nil {
		return err
	}

	// start listening for connections
	s.listener, err = net.Listen("tcp", s.bindAddress)
	if err != nil {
		return err
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

		client := NewClientHandler(conn, s.closeChan, s.db, s.sessionKey)

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
