package game

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/pkg/errors"
	"github.com/yuin/gopher-lua"
	"os"
	"path"
	"strings"
)

// ScriptManager manages game server scripts.
type ScriptManager struct {
	baseDir string
	state   *lua.LState
}

// NewScriptManager creates a new script manager that manages scripts in a baseDir directory.
func NewScriptManager(baseDir string) *ScriptManager {
	sm := &ScriptManager{
		baseDir: baseDir,
		state:   lua.NewState(),
	}

	sm.registerItemModel()
	return sm
}

// Load parses and loads all script files located in the base directory, returning the number of scripts on success. If
// any script fails to load, an error will be returned.
func (s *ScriptManager) Load() (int, error) {
	items, err := os.ReadDir(s.baseDir)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, item := range items {
		if item.IsDir() || !strings.HasSuffix(item.Name(), ".lua") {
			continue
		}

		scriptFile := path.Join(s.baseDir, item.Name())
		err = s.state.DoFile(scriptFile)
		if err != nil {
			return 0, errors.Wrap(err, fmt.Sprintf("unable to load script file: %s", scriptFile))
		}

		count++
	}

	return count, nil
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
