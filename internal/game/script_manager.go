package game

import (
	"fmt"
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
	return &ScriptManager{
		baseDir: baseDir,
		state:   lua.NewState(),
	}
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

func (s *ScriptManager) DoItemEquipped(pe *playerEntity) error {
	err := s.state.CallByParam(lua.P{
		Fn:      s.state.GetGlobal("on_equip"),
		NRet:    0,
		Protect: true,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *ScriptManager) registerPlayerEntityType() {
	mt := s.state.NewTypeMetatable("playerEntity")
	s.state.SetGlobal("playerEntity", mt)
	s.state.SetField(mt, "__index", s.state.SetFuncs(s.state.NewTable(), map[string]lua.LGFunction{
		"username": s.playerEntityGetUsername,
	}))
}

func (s *ScriptManager) playerEntityFromState(l *lua.LState) (*playerEntity, error) {
	ud := l.CheckUserData(1)
	if pe, ok := ud.Value.(*playerEntity); ok {
		return pe, nil
	}

	return nil, fmt.Errorf("expected *playerEntity as first arg")
}

func (s *ScriptManager) playerEntityGetUsername(l *lua.LState) int {
	pe, err := s.playerEntityFromState(l)
	if err != nil {
		return 0
	}

	l.Push(lua.LString(pe.player.Username))
	return 1
}
