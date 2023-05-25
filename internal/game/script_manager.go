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

const (
	combatStatAttackStab int = iota
	combatStatAttackSlash
	combatStatAttackCrush
	combatStatAttackMagic
	combatStatAttackRange
	combatStatDefenseStab
	combatStatDefenseSlash
	combatStatDefenseCrush
	combatStatDefenseMagic
	combatStatDefenseRange
	combatStatStrength
	combatStatPrayer
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
// any script fails to load, an error will be returned. If this method is called after the initial script load is done,
// a full reload of all scripts will be performed. This will destroy the current script manager state, so you need to
// ensure that no consumers are currently awaiting a result from a script.
func (s *ScriptManager) Load() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// clear compiled script cache
	s.protos = nil

	// load all available scripts under the base directory
	_, err := s.loadScriptDirectory(s.baseDir)
	if err != nil {
		return 0, err
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
	return s.doFunction("init_player_tabs", s.playerEntityType(pe, s.state))
}

// DoInterface executes an interface script for an action performed by the player.
func (s *ScriptManager) DoInterface(pe *playerEntity, parent, actor *model.Interface, opCode int) error {
	function := fmt.Sprintf("interface_%d_on_action", parent.ID)
	return s.doFunction(function, s.playerEntityType(pe, s.state), s.interfaceType(actor, s.state), lua.LNumber(opCode))
}

// DoOnEquipItem executes a script to handle a player (un)equipping an item.
func (s *ScriptManager) DoOnEquipItem(pe *playerEntity, item *model.Item) error {
	return s.doFunction("on_equip_item", s.playerEntityType(pe, s.state), s.itemType(item, s.state))
}

// DoOnUnequipItem executes a script to handle a player unequipping an item.
func (s *ScriptManager) DoOnUnequipItem(pe *playerEntity, item *model.Item) error {
	return s.doFunction("on_unequip_item", s.playerEntityType(pe, s.state), s.itemType(item, s.state))
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
		"name": func(state *lua.LState) int {
			item := state.CheckUserData(1).Value.(*model.Item)
			state.Push(lua.LString(item.Name))
			return 1
		},
		"equipment_slot": func(state *lua.LState) int {
			item := state.CheckUserData(1).Value.(*model.Item)
			if item.Attributes == nil {
				state.Push(lua.LNumber(-1))
			} else {
				state.Push(lua.LNumber(item.Attributes.EquipSlotType))
			}

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
		"attack_style": func(state *lua.LState) int {
			pe := state.CheckUserData(1).Value.(*playerEntity)
			weaponStyle := pe.player.EquippedWeaponStyle()

			// if another argument is present on the stack, treat this as a setter
			if state.GetTop() == 2 {
				style := model.AttackStyle(state.CheckInt(2))
				pe.player.SetAttackStyle(weaponStyle, style)
				return 0
			}

			style := pe.player.AttackStyle(weaponStyle)
			state.Push(lua.LNumber(style))

			return 1
		},
		"combat_stats": func(state *lua.LState) int {
			pe := state.CheckUserData(1).Value.(*playerEntity)

			tbl := state.NewTable()
			tbl.RawSetInt(combatStatAttackStab, lua.LNumber(pe.player.CombatStats.Attack.Stab))
			tbl.RawSetInt(combatStatAttackSlash, lua.LNumber(pe.player.CombatStats.Attack.Slash))
			tbl.RawSetInt(combatStatAttackCrush, lua.LNumber(pe.player.CombatStats.Attack.Crush))
			tbl.RawSetInt(combatStatAttackMagic, lua.LNumber(pe.player.CombatStats.Attack.Magic))
			tbl.RawSetInt(combatStatAttackRange, lua.LNumber(pe.player.CombatStats.Attack.Range))

			tbl.RawSetInt(combatStatDefenseStab, lua.LNumber(pe.player.CombatStats.Defense.Stab))
			tbl.RawSetInt(combatStatDefenseSlash, lua.LNumber(pe.player.CombatStats.Defense.Slash))
			tbl.RawSetInt(combatStatDefenseCrush, lua.LNumber(pe.player.CombatStats.Defense.Crush))
			tbl.RawSetInt(combatStatDefenseMagic, lua.LNumber(pe.player.CombatStats.Defense.Magic))
			tbl.RawSetInt(combatStatDefenseRange, lua.LNumber(pe.player.CombatStats.Defense.Range))

			tbl.RawSetInt(combatStatStrength, lua.LNumber(pe.player.CombatStats.Strength))
			tbl.RawSetInt(combatStatPrayer, lua.LNumber(pe.player.CombatStats.Prayer))

			state.Push(tbl)
			return 1
		},
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
		"interface_model": func(state *lua.LState) int {
			pe := state.CheckUserData(1).Value.(*playerEntity)
			interfaceID := state.CheckInt(2)
			itemID := state.CheckInt(3)
			zoom := state.CheckInt(4)

			s.handler.handleSetInterfaceModel(pe, interfaceID, itemID, zoom)
			return 0
		},
		"interface_text": func(state *lua.LState) int {
			pe := state.CheckUserData(1).Value.(*playerEntity)
			interfaceID := state.CheckInt(2)
			text := state.CheckString(3)

			s.handler.handleSetInterfaceText(pe, interfaceID, text)
			return 0
		},
		"interface_setting": func(state *lua.LState) int {
			pe := state.CheckUserData(1).Value.(*playerEntity)
			settingID := state.CheckInt(2)
			value := state.CheckInt(3)

			s.handler.handleSetInterfaceSetting(pe, settingID, value)
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
		"consume_runes": func(state *lua.LState) int {
			pe := state.CheckUserData(1).Value.(*playerEntity)

			// ensure an even number of arguments was given
			stack := state.GetTop() - 1
			if stack%2 != 0 {
				state.ArgError(1, "invalid number of arguments")
				return 0
			}

			// start at the first vararg on the stack
			stackPtr := 2

			// form a slice consisting of rune IDs and amounts
			args := make([]int, stack)
			for i := 0; i < stack; i += 2 {
				args[i] = state.CheckInt(stackPtr)
				args[i+1] = state.CheckInt(stackPtr + 1)
				stackPtr += 2
			}

			// attempt to consume the runes from the player's inventory
			valid := s.handler.handleConsumeRunes(pe, args...)
			state.Push(lua.LBool(valid))
			return 1
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

// loadScriptDirectory parses script files in a directory and compiles them. Subdirectories will be processed as they
// are encountered. If successful, the number of script files loaded will be returned, otherwise an error will be
// returned if any script fails to compile.
func (s *ScriptManager) loadScriptDirectory(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		// recursively examine subdirectories
		if entry.IsDir() {
			subDir := path.Join(dir, entry.Name())
			n, err := s.loadScriptDirectory(subDir)
			if err != nil {
				return 0, err
			}

			count += n
			continue
		}

		// ignore potentially unknown files
		if !strings.HasSuffix(entry.Name(), ".lua") {
			continue
		}

		scriptFile := path.Join(dir, entry.Name())

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

	return count, nil
}
