package game

import (
	"bytes"
	"fmt"
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
const luaTypeItem = "item"

// scriptType enumerates the known types of scripts.
type scriptType int

const (
	scriptTypeInterface scriptType = iota
)

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

// DoInterface executes an interface script for an action performed by the player.
func (s *ScriptManager) DoInterface(pe *playerEntity, interfaceID, opCode int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.state.CallByParam(lua.P{
		Fn:      s.state.GetGlobal(fmt.Sprintf("interface_%d_on_action", interfaceID)),
		NRet:    0,
		Protect: true,
	}, s.playerEntity(pe, s.state), lua.LNumber(opCode))

	if err != nil {
		return err
	}

	return nil
}

// playerEntity creates a Lua user-defined data type for a playerEntity.
func (s *ScriptManager) playerEntity(pe *playerEntity, l *lua.LState) *lua.LUserData {
	ud := l.NewUserData()
	ud.Value = pe
	ud.Metatable = l.GetTypeMetatable(luaTypePlayerEntity)
	return ud
}

// createState creates a new Lua state initialized with user-defined types and compiled functions.
func (s *ScriptManager) createState() (*lua.LState, error) {
	l := lua.NewState()
	s.registerItemModel(l)
	s.registerPlayerModel(l)

	err := s.registerFunctionProtos(l)
	if err != nil {
		return nil, err
	}

	return l, nil
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
	}))
}

// registerItemModel registers metadata for a playerEntity type.
func (s *ScriptManager) registerPlayerModel(l *lua.LState) {
	mt := l.NewTypeMetatable(luaTypePlayerEntity)
	l.SetGlobal(luaTypePlayerEntity, mt)

	l.SetField(mt, "__index", l.SetFuncs(l.NewTable(), map[string]lua.LGFunction{
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
