package server

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/mbpolan/openmcs/internal/game"
	"github.com/mbpolan/openmcs/internal/logger"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/network/request"
	"github.com/mbpolan/openmcs/internal/network/response"
	"github.com/mbpolan/openmcs/internal/store"
	"github.com/pkg/errors"
	"io"
	"net"
	"strings"
	"time"
)

// clientVersion is the client version that is supported by the server.
const clientVersion = 317

// clientState is an enumeration of the various states a player's connection may be in.
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
			c.logDisconnectError(err)
			c.state = failed
		} else {
			c.state = nextState
		}
	}

	// indicate this client handler can be cleaned up
	c.closeChan <- c

	// if the player was added to the game world, remove them and save their persistent data
	if c.player != nil {
		// remove the player from the game world
		c.game.RemovePlayer(c.player)

		err := c.store.SavePlayer(c.player)
		if err != nil {
			logger.Errorf("failed to save player %d: %s", c.player.ID, err)
		}
	}
}

// logDisconnectError possibly logs information about the player disconnecting.
func (c *ClientHandler) logDisconnectError(err error) {
	// if the underlying cause is an eof, don't log an error since that indicates the client disconnected from us
	cause1 := errors.Unwrap(err)
	cause2 := errors.Unwrap(cause1)
	if cause1 == io.EOF || cause2 == io.EOF {
		username := "(unknown)"
		if c.player != nil && c.player.Username != "" {
			username = c.player.Username
		}

		logger.Infof("disconnecting player: %s", username)
		return
	}

	logger.Errorf("disconnecting player due to error: %s", err)
}

func (c *ClientHandler) handleInitialization() (clientState, error) {
	header, err := c.reader.Peek()
	if err != nil {
		return failed, errors.Wrap(err, "failed to read init packet header")
	}

	// expect an init request first
	if header != request.InitRequestHeader {
		return failed, fmt.Errorf("unexpected init packet header: %2x", header)
	}

	// read the contents of the init request
	var req request.InitRequest
	err = req.Read(c.reader)
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
	header, err := c.reader.Peek()
	if err != nil {
		return failed, errors.Wrap(err, "failed to read login packet header")
	}

	// expect a login request (either a reconnect attempt or a new connection)
	if header != request.ReconnectLoginRequestHeader && header != request.NewLoginRequestHeader {
		return failed, fmt.Errorf("unexpected login packet header: %2x", header)
	}

	// read the contents of the login request
	var req request.LoginRequest
	err = req.Read(c.reader)
	if err != nil {
		return failed, errors.Wrap(err, "unexpected login request contents")
	}

	// validate if the client is supported by the server
	if req.Version != clientVersion {
		resp := response.NewFailedInitResponse(response.InitGameUpdated)
		err := resp.Write(c.writer)
		return failed, err
	}

	// load the player's data, if it exists
	player, err := c.store.LoadPlayer(req.Username)
	if err != nil {
		// fall through
		logger.Errorf("failed to load player %s: %s", req.Username, err)
		player = nil
	}

	// does a player with that username even exist?
	if player == nil {
		resp := response.NewFailedInitResponse(response.InitInvalidUsername)
		err := resp.Write(c.writer)
		return failed, err
	}

	// hash their password for comparison
	passwordHash := c.hashPassword(req.Password)
	if player.PasswordHash != passwordHash {
		// TODO: track this as a failed login attempt
		resp := response.NewFailedInitResponse(response.InitInvalidUsername)
		err := resp.Write(c.writer)
		return failed, err
	}

	// check if the player can be added to the game
	result := c.game.ValidatePlayer(player)
	if result != game.ValidationResultSuccess {
		var resp response.Response

		switch result {
		case game.ValidationResultAlreadyLoggedIn:
			resp = response.NewFailedInitResponse(response.InitAccountLoggedIn)
		case game.ValidationResultNoCapacity:
			resp = response.NewFailedInitResponse(response.InitServerFull)
		default:
			break
		}

		if resp != nil {
			err = resp.Write(c.writer)
		}

		return failed, err
	}

	// the player has now authenticated and can be added to the game
	c.player = player

	// send a confirmation to the client
	resp := response.NewLoggedInInitResponse(c.player.Type, c.player.Flagged)
	err = resp.Write(c.writer)
	if err != nil {
		return failed, errors.Wrap(err, "failed to send logged in response")
	}

	// add the player to the game world
	c.game.AddPlayer(player, c.writer)

	logger.Infof("connected new player: %s (%s)", c.player.Username, c.conn.RemoteAddr().String())
	return active, nil
}

// hashPassword computes a hash of the player's password.
func (c *ClientHandler) hashPassword(password string) string {
	// use a sha512/256 hash algorithm for passwords
	hash := sha512.Sum512_256([]byte(password))
	return strings.ToLower(hex.EncodeToString(hash[:]))
}

func (c *ClientHandler) handleLoop() (clientState, error) {
	header, err := c.reader.Peek()
	if err != nil {
		return failed, errors.Wrap(err, "unexpected error while waiting for packet header")
	}

	// maintain current state
	var nextState = c.state

	switch header {
	case request.KeepAliveRequestHeader:
		// idle/keep-alive
		var req request.KeepAliveRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.lastHeartbeat = time.Now()

	case request.FocusRequestHeader:
		// client window focus has changed
		var req request.FocusChangeRequest
		err = req.Read(c.reader)

	case request.ClientClickRequestHeader:
		// the player clicked somewhere on the client window
		var req request.ClientClickRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.MarkPlayerActive(c.player)

	case request.RegionChangeRequestHeader:
		// the player entered a new map region
		var req request.RegionChangeRequest
		err = req.Read(c.reader)

	case request.CameraModeRequestHeader:
		// the player moved their client's camera
		var req request.CameraModeRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.MarkPlayerActive(c.player)

	case request.RegionLoadedRequestHeader:
		// the player's client finished loading a new map region
		var req request.RegionLoadedRequest
		err = req.Read(c.reader)

	case request.ReportRequestHeader:
		// the player sent an abuse report
		var req request.ReportRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.ProcessAbuseReport(c.player, req.Username, req.Reason, req.EnableMute)

	case request.CloseInterfaceRequestHeader:
		// the player's client dismissed the current interface, if any
		var req request.CloseInterfaceRequest
		err = req.Read(c.reader)

	case request.PlayerIdleRequestHeader:
		// the player has become idle
		var req request.PlayerIdleRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.MarkPlayerInactive(c.player)

	case request.PlayerChatRequestHeader:
		// the player sent a chat message
		var req request.PlayerChatRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.DoPlayerChat(c.player, req.Effect, req.Color, req.Text)

	case request.ChatCommandRequestHeader:
		// the player sent a chat command
		var req request.ChatCommandRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.DoPlayerChatCommand(c.player, req.Text)

	case request.PrivateChatRequestHeader:
		// the player sent a private chat message
		var req request.PrivateChatRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.DoPlayerPrivateChat(c.player, req.Recipient, req.Text)

	case request.ChangeModesRequestHeader:
		// the player changed one or more chat or interaction modes
		var req request.ChangeModesRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.SetPlayerModes(c.player, req.PublicChat, req.PrivateChat, req.Interaction)

	case request.WalkRequestHeader, request.WalkOnCommandRequestHeader, request.WalkMinimap:
		// the player started walking to a destination on the map
		var req request.WalkRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.WalkPlayer(c.player, req.Start, req.Waypoints)

	case request.TakeGroundItemRequestHeader:
		// the player tried to pick up a ground item
		var req request.TakeGroundItemRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.DoTakeGroundItem(c.player, req.ItemID, req.GlobalPos)

	case request.DropInventoryItemRequestHeader:
		// the player dropped an inventory item
		var req request.DropInventoryItemRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.DoDropInventoryItem(c.player, req.ItemID, req.InterfaceID, req.SecondaryActionID)

	case request.SwapInventoryItemRequestHeader:
		// the player rearranged an item in their inventory
		var req request.SwapInventoryItemRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.DoSwapInventoryItem(c.player, req.FromSlot, req.ToSlot, req.InterfaceID)

	case request.EquipItemRequestHeader:
		// the player equipped an item from their inventory
		var req request.EquipItemRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.DoEquipItem(c.player, req.ItemID, req.InterfaceID, req.SecondaryActionID)

	case request.UnequipItemRequestHeader:
		// the player unequipped an item from their equipment
		var req request.UnequipItemRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.DoUnequipItem(c.player, req.ItemID, req.InterfaceID, req.SlotType)

	case request.UseItemRequestHeader:
		// the player initiated the default action on an item
		var req request.UseItemRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.DoUseItem(c.player, req.ItemID, req.InterfaceID, req.ActionID)

	case request.UseInventoryItemsRequestHeader:
		// the player used an inventory item on another
		var req request.UseInventoryItemsRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.DoUseInventoryItem(c.player, req.SourceItemID, req.SourceInterfaceID, req.SourceSlotID,
			req.TargetItemID, req.TargetInterfaceID, req.TargetSlotID)

	case request.AddFriendRequestHeader:
		// the player requested another player be added to their friends list
		var req request.ModifyFriendRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.AddFriend(c.player, req.Username)

	case request.RemoveFriendRequestHeader:
		// the player requested another player be removed from their friends list
		var req request.ModifyFriendRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.RemoveFriend(c.player, req.Username)

	case request.AddIgnoreRequestHeader:
		// the player requested another player be added to their ignore list
		var req request.ModifyIgnoreRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.AddIgnored(c.player, req.Username)

	case request.RemoveIgnoreRequestHeader:
		// the player requested another player be removed from their ignore list
		var req request.ModifyIgnoreRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.RemoveIgnored(c.player, req.Username)

	case request.InterfaceActionRequestHeader:
		// the player has performed an action on an interface
		var req request.InterfaceActionRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.DoInterfaceAction(c.player, req.Action)

	case request.InteractObjectRequestHeader:
		// the player interacted with an object
		var req request.InteractObjectRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.DoInteractWithObject(c.player, req.Action, req.GlobalPos)

	case request.CastSpellOnItemRequestHeader:
		// the player cast a spell on an inventory item
		var req request.CastSpellOnItemRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.DoCastSpellOnItem(c.player, req.SlotID, req.ItemID, req.InventoryInterfaceID, req.SpellInterfaceID)

	case request.CharacterDesignRequestHeader:
		// the player submitted a new character design
		var req request.CharacterDesignRequest
		err = req.Read(c.reader)
		if err != nil {
			break
		}

		c.game.DoSetPlayerDesign(c.player, req.Gender, req.Base, req.BodyColors)

	default:
		// unknown packet
		err = fmt.Errorf("unexpected packet header: %2x", header)
	}

	if err != nil {
		return failed, err
	}

	return nextState, nil
}
