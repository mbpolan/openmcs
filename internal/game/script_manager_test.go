package game

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ScriptManager_Load(t *testing.T) {
	sm := NewScriptManager("../../scripts")

	_, err := sm.Load()
	assert.NoError(t, err)

	pe := &playerEntity{player: &model.Player{Username: "mike"}}
	item := &model.Item{ID: 42}
	err = sm.DoItemEquipped(pe, item)
	assert.NoError(t, err)
}
