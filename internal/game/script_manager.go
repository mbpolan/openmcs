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
	"strconv"
	"strings"
	"sync"
)

// scriptType enumerates the known types of scripts.
type scriptType int

const (
	scriptTypeInterface scriptType = iota
)

// ScriptManager manages game server scripts.
type ScriptManager struct {
	baseDir    string
	interfaces map[int]*lua.FunctionProto
	state      *lua.LState
	mu         sync.Mutex
}

// NewScriptManager creates a new script manager that manages scripts in a baseDir directory.
func NewScriptManager(baseDir string) *ScriptManager {
	sm := &ScriptManager{
		baseDir:    baseDir,
		interfaces: map[int]*lua.FunctionProto{},
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

	count := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".lua") {
			continue
		}

		scriptFile := path.Join(s.baseDir, entry.Name())

		// determine the type of script based on its filename
		components := strings.Split(entry.Name(), "_")
		if len(components) < 2 {
			return 0, fmt.Errorf("invalid script file name: %s", scriptFile)
		}

		var sType scriptType
		var interfaceID int
		switch strings.ToLower(components[0]) {
		case "inf":
			sType = scriptTypeInterface
			interfaceID, err = strconv.Atoi(components[1])
			if err != nil {
				return 0, fmt.Errorf("invalid file name for interface script: %s", scriptFile)
			}
		default:
			return 0, fmt.Errorf("unknown script type: %s", scriptFile)
		}

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

		switch sType {
		case scriptTypeInterface:
			s.interfaces[interfaceID] = compiled
		}

		count++
	}

	// create an initial state
	s.state, err = s.createState()
	if err != nil {
		return 0, err
	}

	return count, nil
}

// DoInterface executes an interface script for an action performed by the player.
func (s *ScriptManager) DoInterface(interfaceID, opCode int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.state.CallByParam(lua.P{
		Fn:      s.state.GetGlobal(fmt.Sprintf("interface_on_action_%d", interfaceID)),
		NRet:    0,
		Protect: true,
	}, lua.LNumber(opCode))

	if err != nil {
		return err
	}

	return nil
}

func (s *ScriptManager) DoItemEquipped(pe *playerEntity, item *model.Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	ud := s.state.NewUserData()
	ud.Value = item
	ud.Metatable = s.state.GetTypeMetatable("item")

	err := s.state.CallByParam(lua.P{
		Fn:      s.state.GetGlobal("on_equip"),
		NRet:    0,
		Protect: true,
	}, ud)

	if err != nil {
		return err
	}

	return nil
}

// createState creates a new Lua state initialized with user-defined types and compiled functions.
func (s *ScriptManager) createState() (*lua.LState, error) {
	l := lua.NewState()
	s.registerItemModel(l)

	err := s.registerFunctionProtos(l)
	if err != nil {
		return nil, err
	}

	return l, nil
}

// registerItemModel registers metadata for a model.Item type.
func (s *ScriptManager) registerItemModel(l *lua.LState) {
	mt := l.NewTypeMetatable("item")
	l.SetGlobal("item", mt)

	l.SetField(mt, "__index", l.SetFuncs(l.NewTable(), map[string]lua.LGFunction{
		"id": func(state *lua.LState) int {
			item := state.CheckUserData(1).Value.(*model.Item)
			state.Push(lua.LNumber(item.ID))
			return 1
		},
	}))
}

// registerFunctionProtos executes compiled functions into a Lua state.
func (s *ScriptManager) registerFunctionProtos(l *lua.LState) error {
	for id, proto := range s.interfaces {
		l.Push(l.NewFunctionFromProto(proto))

		err := l.PCall(0, lua.MultRet, nil)
		if err != nil {
			return errors.Wrapf(err, "failed to execute function proto for interface %d", id)
		}
	}

	return nil
}