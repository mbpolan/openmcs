package game

import (
	"bytes"
	"fmt"
	"github.com/mbpolan/openmcs/internal/logger"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/pkg/errors"
	"github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
	"os"
	"path"
	"strings"
	"sync"
)

const luaTypePlayerEntity = "playerEntity"
const luaTypeInterface = "interface"
const luaTypeItem = "item"

// ScriptManager manages game server scripts.
type ScriptManager struct {
	baseDir string
	handler ScriptHandler
	protos  []*lua.FunctionProto
	state   *lua.LState
	mu      sync.Mutex
}

// NewScriptManager creates a new script manager that manages scripts in a baseDir directory.
func NewScriptManager(baseDir string, handler ScriptHandler) *ScriptManager {
	sm := &ScriptManager{
		baseDir: baseDir,
		handler: handler,
	}

	return sm
}

// Load parses and loads all script files located in the base directory, returning the number of scripts on success. If
// any script fails to load, an error will be returned.
func (s *ScriptManager) Load() (int, error) {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".lua") {
			continue
		}

		scriptFile := path.Join(s.baseDir, entry.Name())

		// read the contents of the script
		data, err := os.ReadFile(scriptFile)
		if err != nil {
			return 0, errors.Wrapf(err, "failed to open script file: %s", scriptFile)
		}

		// parse it into a lua chunk
		chunk, err := parse.Parse(bytes.NewReader(data), scriptFile)
		if err != nil {
			return 0, errors.Wrapf(err, "failed to compile script file: %s", scriptFile)
		}

		// compile the script into a function proto
		compiled, err := lua.Compile(chunk, scriptFile)
		if err != nil {
			return 0, errors.Wrap(err, fmt.Sprintf("unable to load script file: %s", scriptFile))
		}

		s.protos = append(s.protos, compiled)
	}

	// create an initial state
	s.state, err = s.createState()
	if err != nil {
		return 0, err
	}

	return len(s.protos), nil
}

// DoPlayerInit executes a script to initialize a player when they join the game.
func (s *ScriptManager) DoPlayerInit(pe *playerEntity) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// attempt to call a function for this interface's handler
	function := "init_player_tabs"
	err := s.state.CallByParam(lua.P{
		Fn:      s.state.GetGlobal(function),
		NRet:    0,
		Protect: true,
	}, s.playerEntityType(pe, s.state))

	return s.checkResult(function, err)

}

// DoInterface executes an interface script for an action performed by the player.
func (s *ScriptManager) DoInterface(pe *playerEntity, parent, actor *model.Interface, opCode int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// attempt to call a function for this interface's handler
	function := fmt.Sprintf("interface_%d_on_action", parent.ID)
	err := s.state.CallByParam(lua.P{
		Fn:      s.state.GetGlobal(function),
		NRet:    0,
		Protect: true,
	}, s.playerEntityType(pe, s.state), s.interfaceType(actor, s.state), lua.LNumber(opCode))

	return s.checkResult(function, err)
}

// DoOnEquipItem executes a script to handle a player (un)equipping an item.
func (s *ScriptManager) DoOnEquipItem(pe *playerEntity, item *model.Item) error {
	return s.doFunction("on_equip_item", s.playerEntityType(pe, s.state), s.itemType(item, s.state))
}

// DoOnUnequipItem executes a script to handle a player unequipping an item.
func (s *ScriptManager) DoOnUnequipItem(pe *playerEntity, item *model.Item) error {
	return s.doFunction("on_unequip_item", s.playerEntityType(pe, s.state))
}

// playerEntity creates a Lua user-defined data type for a playerEntity.
func (s *ScriptManager) playerEntityType(pe *playerEntity, l *lua.LState) *lua.LUserData {
	ud := l.NewUserData()
	ud.Value = pe
	ud.Metatable = l.GetTypeMetatable(luaTypePlayerEntity)
	return ud
}

// itemType creates a Lua user-defined data type for a model.Item.
func (s *ScriptManager) itemType(item *model.Item, l *lua.LState) *lua.LUserData {
	ud := l.NewUserData()
	ud.Value = item
	ud.Metatable = l.GetTypeMetatable(luaTypeItem)
	return ud
}

// interfaceType creates a Lua user-defined data type for a model.Interface.
func (s *ScriptManager) interfaceType(inf *model.Interface, l *lua.LState) *lua.LUserData {
	ud := l.NewUserData()
	ud.Value = inf
	ud.Metatable = l.GetTypeMetatable(luaTypeInterface)
	return ud
}

// createState creates a new Lua state initialized with user-defined types and compiled functions.
func (s *ScriptManager) createState() (*lua.LState, error) {
	l := lua.NewState()
	s.registerInterfaceModel(l)
	s.registerItemModel(l)
	s.registerPlayerModel(l)

	err := s.registerFunctionProtos(l)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// doFunction executes a function in the Lua state.
func (s *ScriptManager) doFunction(function string, params ...lua.LValue) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.state.CallByParam(lua.P{
		Fn:      s.state.GetGlobal(function),
		NRet:    0,
		Protect: true,
	}, params...)

	return s.checkResult(function, err)
}

// checkResult inspects an error returned by the Lua VM and includes additional logging.
func (s *ScriptManager) checkResult(function string, err error) error {
	if err != nil {
		if le, ok := err.(*lua.ApiError); ok {
			logger.Errorf("lua script error on calling function %s\nerror: %s\nstack:\n%s", function, le.Object, le.StackTrace)
		}
		return err
	}

	return nil
}

// registerItemModel registers metadata for a model.Interface type.
func (s *ScriptManager) registerInterfaceModel(l *lua.LState) {
	mt := l.NewTypeMetatable(luaTypeInterface)
	l.SetGlobal(luaTypeInterface, mt)

	l.SetField(mt, "__index", l.SetFuncs(l.NewTable(), map[string]lua.LGFunction{
		"id": func(state *lua.LState) int {
			inf := state.CheckUserData(1).Value.(*model.Interface)
			state.Push(lua.LNumber(inf.ID))
			return 1
		},
	}))
}

// registerItemModel registers metadata for a model.Item type.
func (s *ScriptManager) registerItemModel(l *lua.LState) {
	mt := l.NewTypeMetatable(luaTypeItem)
	l.SetGlobal(luaTypeItem, mt)

	l.SetField(mt, "__index", l.SetFuncs(l.NewTable(), map[string]lua.LGFunction{
		"id": func(state *lua.LState) int {
			item := state.CheckUserData(1).Value.(*model.Item)
			state.Push(lua.LNumber(item.ID))
			return 1
		},
		"weapon_style": func(state *lua.LState) int {
			item := state.CheckUserData(1).Value.(*model.Item)
			if item.Attributes == nil {
				state.Push(lua.LNumber(-1))
			} else {
				state.Push(lua.LNumber(item.Attributes.WeaponStyle))
			}

			return 1
		},
	}))
}

// registerItemModel registers metadata for a playerEntity type.
func (s *ScriptManager) registerPlayerModel(l *lua.LState) {
	mt := l.NewTypeMetatable(luaTypePlayerEntity)
	l.SetGlobal(luaTypePlayerEntity, mt)

	l.SetField(mt, "__index", l.SetFuncs(l.NewTable(), map[string]lua.LGFunction{
		"sidebar_clear": func(state *lua.LState) int {
			pe := state.CheckUserData(1).Value.(*playerEntity)
			sidebarID := state.CheckInt(2)

			s.handler.handleClearSidebarInterface(pe, sidebarID)
			return 0
		},
		"sidebar_interface": func(state *lua.LState) int {
			pe := state.CheckUserData(1).Value.(*playerEntity)
			sidebarID := state.CheckInt(2)
			interfaceID := state.CheckInt(3)

			s.handler.handleSetSidebarInterface(pe, interfaceID, sidebarID)
			return 0
		},
		"interface_text": func(state *lua.LState) int {
			pe := state.CheckUserData(1).Value.(*playerEntity)
			interfaceID := state.CheckInt(2)
			text := state.CheckString(3)

			s.handler.handleSetInterfaceText(pe, interfaceID, text)
			return 0
		},
		"equipped_item": func(state *lua.LState) int {
			pe := state.CheckUserData(1).Value.(*playerEntity)
			slotType := state.CheckInt(2)

			slot := pe.player.EquipmentSlot(model.EquipmentSlotType(slotType))
			if slot == nil {
				state.Push(lua.LNil)
			} else {
				state.Push(s.itemType(slot.Item, state))
			}

			return 1
		},
		"disconnect": func(state *lua.LState) int {
			pe := state.CheckUserData(1).Value.(*playerEntity)
			s.handler.handleRemovePlayer(pe)
			return 0
		},
	}))
}

// registerFunctionProtos executes compiled functions into a Lua state.
func (s *ScriptManager) registerFunctionProtos(l *lua.LState) error {
	for _, proto := range s.protos {
		l.Push(l.NewFunctionFromProto(proto))

		err := l.PCall(0, lua.MultRet, nil)
		if err != nil {
			return errors.Wrapf(err, "failed to execute function proto")
		}
	}

	return nil
}
