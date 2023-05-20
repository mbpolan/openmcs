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
}

// NewScriptManager creates a new script manager that manages scripts in a baseDir directory.
func NewScriptManager(baseDir string) *ScriptManager {
	sm := &ScriptManager{
		baseDir:    baseDir,
		interfaces: map[int]*lua.FunctionProto{},
		state:      lua.NewState(),
	}

	sm.registerItemModel()
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

	return count, nil
}

// DoInterface executes an interface script for an action performed by the player.
func (s *ScriptManager) DoInterface(interfaceID, opCode int) error {
	err := s.state.CallByParam(lua.P{
		Fn:      s.state.GetGlobal(fmt.Sprintf("on_action_%d", interfaceID)),
		NRet:    0,
		Protect: true,
	}, lua.LNumber(opCode))

	if err != nil {
		return err
	}

	return nil
}

func (s *ScriptManager) DoItemEquipped(pe *playerEntity, item *model.Item) error {
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

func (s *ScriptManager) registerItemModel() {
	mt := s.state.NewTypeMetatable("item")
	s.state.SetGlobal("item", mt)

	s.state.SetField(mt, "__index", s.state.SetFuncs(s.state.NewTable(), map[string]lua.LGFunction{
		"id": func(state *lua.LState) int {
			item := state.CheckUserData(1).Value.(*model.Item)
			state.Push(lua.LNumber(item.ID))
			return 1
		},
	}))
}
