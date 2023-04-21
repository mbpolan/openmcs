package server

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/game"
	"github.com/mbpolan/openmcs/internal/logger"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/network/request"
	"github.com/mbpolan/openmcs/internal/network/response"
	"github.com/mbpolan/openmcs/internal/store"
	"github.com/pkg/errors"
	"net"
	"time"
)

type clientState int

const (
	initializing clientState = iota
	loggingIn
	active
	failed
)

// ClientHandler is responsible for managing the state and communications for a single client.
type ClientHandler struct {
	conn          net.Conn
	game          *game.Game
	reader        *network.ProtocolReader
	writer        *network.ProtocolWriter
	closeChan     chan *ClientHandler
	lastHeartbeat time.Time
	player        *model.Player
	store         *store.Store
	sessionKey    uint64
	state         clientState
}

// NewClientHandler returns a new handler for a client connection. When the handler terminates, it will write to the
// provided closeChan to indicate its work is complete.
func NewClientHandler(conn net.Conn, closeChan chan *ClientHandler, store *store.Store, game *game.Game, sessionKey uint64) *ClientHandler {
	return &ClientHandler{
		conn:       conn,
		game:       game,
		store:      store,
		reader:     network.NewReader(conn),
		writer:     network.NewWriter(conn),
		closeChan:  closeChan,
		state:      initializing,
		sessionKey: sessionKey,
	}
}

// Handle processes request for the client connection.
func (c *ClientHandler) Handle() {
	run := true
	defer c.conn.Close()

	// continually process request from the client until we reach either a graceful close or error state
	for run {
		var nextState clientState
		var err error

		switch c.state {
		case initializing:
			nextState, err = c.handleInitialization()
		case loggingIn:
			nextState, err = c.handleLogin()
		case active:
			nextState, err = c.handleLoop()
		case failed:
			run = false
		}

		if err != nil {
			logger.Errorf("disconnecting player due to error: %s", err)
			c.state = failed
		} else {
			c.state = nextState
		}
	}

	// remove the player from the game world
	c.game.RemovePlayer(c.player)

	// indicate this client handler can be cleaned up
	c.closeChan <- c
}

func (c *ClientHandler) handleInitialization() (clientState, error) {
	b, err := c.reader.Uint8()
	if err != nil {
		return failed, errors.Wrap(err, "failed to read init packet header")
	}

	// expect an init request first
	if b != request.InitRequestHeader {
		return failed, fmt.Errorf("unexpected init packet header: %2x", b)
	}

	// read the contents of the init request
	_, err = request.ReadInitRequest(c.reader)
	if err != nil {
		return failed, errors.Wrap(err, "unexpected login packet contents")
	}

	// write some padding bytes (ignored by client)
	padding := response.NewBlankResponse(8)
	err = padding.Write(c.writer)
	if err != nil {
		return failed, errors.Wrap(err, "failed to send padding")
	}

	// accept the session
	resp := response.NewAcceptedInitResponse(c.sessionKey)
	err = resp.Write(c.writer)
	if err != nil {
		return failed, errors.Wrap(err, "failed to send init response")
	}

	return loggingIn, nil
}

func (c *ClientHandler) handleLogin() (clientState, error) {
	b, err := c.reader.Uint8()
	if err != nil {
		return failed, errors.Wrap(err, "failed to read login packet header")
	}

	// expect a login request (either a reconnect attempt or a new connection)
	if b != request.ReconnectLoginRequestHeader && b != request.NewLoginRequestHeader {
		return failed, fmt.Errorf("unexpected login packet header: %2x", b)
	}

	// read the contents of the login request
	req, err := request.ReadLoginRequest(c.reader)
	if err != nil {
		return failed, errors.Wrap(err, "unexpected login request contents")
	}

	// load the player's data, if it exists
	c.player, err = c.store.LoadPlayer(req.Username)

	// authenticate the player
	if c.player == nil || c.player.Password != req.Password {
		resp := response.NewFailedInitResponse(response.InitInvalidUsername)
		err := resp.Write(c.writer)
		return failed, err
	}

	// send a confirmation to the client
	resp := response.NewLoggedInInitResponse(c.player.Type, c.player.Flagged)
	err = resp.Write(c.writer)
	if err != nil {
		return failed, errors.Wrap(err, "failed to send logged in response")
	}

	// add the player to the game world
	c.game.AddPlayer(c.player, c.writer)
	logger.Infof("connected new player: %s", c.player.Username)

	return active, nil
}

func (c *ClientHandler) handleLoop() (clientState, error) {
	b, err := c.reader.Uint8()
	if err != nil {
		return failed, errors.Wrap(err, "unexpected error while waiting for packet header")
	}

	// maintain current state
	var nextState = c.state

	switch b {
	case request.KeepAliveRequestHeader:
		// idle/keep-alive
		c.lastHeartbeat = time.Now()

	case request.FocusRequestHeader:
		// client window focus has changed
		_, err = request.ReadFocusRequest(c.reader)

	case request.ClientClickRequestHeader:
		// the player clicked somewhere on the client window
		_, err = request.ReadClientClickRequest(c.reader)
		c.game.MarkPlayerActive(c.player)

	case request.RegionChangeRequestHeader:
		// the player entered a new map region
		_, err = request.ReadRegionChangeRequest(c.reader)

	case request.CameraModeRequestHeader:
		// the player moved their client's camera
		_, err = request.ReadCameraModeRequest(c.reader)
		c.game.MarkPlayerActive(c.player)

	case request.RegionLoadedRequestHeader:
		// the player's client finished loading a new map region

	case request.ReportRequestHeader:
		// the player sent an abuse report
		req, err := request.ReadReportRequest(c.reader)
		if err == nil {
			c.game.ProcessAbuseReport(c.player, req.Username, req.Reason, req.EnableMute)
		}

	case request.CloseInterfaceRequestHeader:
		// the player's client dismissed the current interface, if any

	case request.PlayerIdleRequestHeader:
		// the player has become idle
		c.game.MarkPlayerInactive(c.player)

	case request.PlayerChatRequestHeader:
		// the player sent a chat message
		req, err := request.ReadPlayerChatRequest(c.reader)
		if err == nil {
			c.game.DoPlayerChat(c.player, req.Effect, req.Color, req.Text)
		}

	case request.PrivateChatRequestHeader:
		// the player sent a private chat message
		req, err := request.ReadPrivateChatRequest(c.reader)
		if err == nil {
			c.game.DoPlayerPrivateChat(c.player, req.Recipient, req.Text)
		}

	case request.ChangeModesRequestHeader:
		// the player changed one or more chat or interaction modes
		req, err := request.ReadChangeModesRequest(c.reader)
		if err == nil {
			c.game.SetPlayerModes(c.player, req.PublicChat, req.PrivateChat, req.Interaction)
		}

	case request.WalkRequestHeader, request.WalkOnCommandRequestHeader:
		// the player started walking to a destination on the map
		req, err := request.ReadWalkRequest(c.reader)
		if err == nil {
			c.game.WalkPlayer(c.player, req.Start, req.Waypoints)
		}

	case request.AddFriendRequestHeader:
		// the player requested another player be added to their friends list
		req, err := request.ReadModifyFriendRequest(c.reader)
		if err == nil {
			c.game.AddFriend(c.player, req.Username)
		}

	case request.RemoveFriendRequestHeader:
		// the player requested another player be removed from their friends list
		req, err := request.ReadModifyFriendRequest(c.reader)
		if err == nil {
			c.game.RemoveFriend(c.player, req.Username)
		}

	case request.AddIgnoreRequestHeader:
		// the player requested another player be added to their ignore list
		req, err := request.ReadModifyIgnoreRequest(c.reader)
		if err == nil {
			c.game.AddIgnored(c.player, req.Username)
		}

	case request.RemoveIgnoreRequestHeader:
		// the player requested another player be removed from their ignore list
		req, err := request.ReadModifyIgnoreRequest(c.reader)
		if err == nil {
			c.game.RemoveIgnored(c.player, req.Username)
		}

	case request.LogoutRequestHeader:
		// the player has requested to log out
		req, err := request.ReadLogoutRequest(c.reader)
		if err == nil {
			c.game.RequestLogout(c.player, req.Action)
		}

	default:
		// unknown packet
		err = fmt.Errorf("unexpected packet header: %2x", b)
	}

	if err != nil {
		return failed, err
	}

	return nextState, nil
}
