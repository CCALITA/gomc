//go:build !ci

package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateManager_Initial(t *testing.T) {
	sm := NewStateManager()
	assert.Equal(t, GameStateMainMenu, sm.CurrentState())
}

func TestStateManager_SetState(t *testing.T) {
	sm := NewStateManager()
	sm.SetState(GameStatePlaying)
	assert.Equal(t, GameStatePlaying, sm.CurrentState())
}

func TestStateManager_SameState_NoOp(t *testing.T) {
	sm := NewStateManager()
	entered := false
	sm.OnEnter(GameStateMainMenu, func() { entered = true })
	sm.SetState(GameStateMainMenu)
	assert.False(t, entered)
}

func TestStateManager_Callbacks(t *testing.T) {
	sm := NewStateManager()
	var log []string

	sm.OnExit(GameStateMainMenu, func() { log = append(log, "exit-menu") })
	sm.OnEnter(GameStatePlaying, func() { log = append(log, "enter-playing") })
	sm.OnExit(GameStatePlaying, func() { log = append(log, "exit-playing") })
	sm.OnEnter(GameStatePaused, func() { log = append(log, "enter-paused") })

	sm.SetState(GameStatePlaying)
	assert.Equal(t, []string{"exit-menu", "enter-playing"}, log)

	sm.SetState(GameStatePaused)
	assert.Equal(t, []string{"exit-menu", "enter-playing", "exit-playing", "enter-paused"}, log)
}

func TestStateConstants(t *testing.T) {
	assert.Equal(t, State(0), GameStateMainMenu)
	assert.Equal(t, State(1), GameStatePlaying)
	assert.Equal(t, State(2), GameStatePaused)
}

func TestTickRate(t *testing.T) {
	assert.InDelta(t, 0.05, tickInterval, 0.001)
}
