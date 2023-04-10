package server

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/logger"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/network/requests"
	"github.com/mbpolan/openmcs/internal/network/responses"
	"github.com/pkg/errors"
	"net"
)

type clientState int

const (
	initializing clientState = iota
	failed
)

// ClientHandler is responsible for managing the state and communications for a single client.
type ClientHandler struct {
	conn      net.Conn
	reader    *network.ProtocolReader
	writer    *network.ProtocolWriter
	closeChan chan *ClientHandler
	state     clientState
}

// NewClientHandler returns a new handler for a client connection. When the handler terminates, it will write to the
// provided closeChan to indicate its work is complete.
func NewClientHandler(conn net.Conn, closeChan chan *ClientHandler) *ClientHandler {
	return &ClientHandler{
		conn:      conn,
		reader:    network.NewReader(conn),
		writer:    network.NewWriter(conn),
		closeChan: closeChan,
		state:     initializing,
	}
}

// Handle processes requests for the client connection.
func (c *ClientHandler) Handle() {
	run := true

	// continually process requests from the client until we reach either a graceful close or error state
	for run {
		switch c.state {
		case initializing:
			nextState, err := c.handleInitialization()
			if err != nil {
				logger.Errorf("error during client initialization: %s", err)
				run = false
			}

			c.state = nextState
		case failed:
			run = false
		}
	}

	// indicate this client handler can be cleaned up
	c.closeChan <- c
}

func (c *ClientHandler) handleInitialization() (clientState, error) {
	b, err := c.reader.Byte()
	if err != nil {
		return failed, errors.Wrap(err, "failed to read init packet header")
	}

	// expect an init request first
	if b != requests.InitRequestHeader {
		return failed, fmt.Errorf("unexpected init packet header: %2x", b)
	}

	// read the contents of the init request
	_, err = requests.ReadInitRequest(c.reader)
	if err != nil {
		return failed, fmt.Errorf("unexpected login packet contents: %s", err)
	}

	// TODO: handle the init request properly
	resp := responses.NewFailedInitResponse(responses.InitInvalidUsername)
	err = resp.Write(c.writer)
	if err != nil {
		return failed, errors.Wrap(err, "failed to write init response")
	}

	return failed, nil
}
