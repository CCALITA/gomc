//go:build !ci

package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateManager_Initial(t *testing.T) {
	sm := NewStateManager()
	assert.Equal(t, StateMainMenu, sm.CurrentState())
}

func TestStateManager_SetState(t *testing.T) {
	sm := NewStateManager()
	sm.SetState(StatePlaying)
	assert.Equal(t, StatePlaying, sm.CurrentState())
}

func TestStateManager_SameState_NoOp(t *testing.T) {
	sm := NewStateManager()
	entered := false
	sm.OnEnter(StateMainMenu, func() { entered = true })
	sm.SetState(StateMainMenu)
	assert.False(t, entered)
}

func TestStateManager_Callbacks(t *testing.T) {
	sm := NewStateManager()
	var log []string

	sm.OnExit(StateMainMenu, func() { log = append(log, "exit-menu") })
	sm.OnEnter(StatePlaying, func() { log = append(log, "enter-playing") })
	sm.OnExit(StatePlaying, func() { log = append(log, "exit-playing") })
	sm.OnEnter(StatePaused, func() { log = append(log, "enter-paused") })

	sm.SetState(StatePlaying)
	assert.Equal(t, []string{"exit-menu", "enter-playing"}, log)

	sm.SetState(StatePaused)
	assert.Equal(t, []string{"exit-menu", "enter-playing", "exit-playing", "enter-paused"}, log)
}

func TestStateConstants(t *testing.T) {
	assert.Equal(t, State(0), StateMainMenu)
	assert.Equal(t, State(1), StatePlaying)
	assert.Equal(t, State(2), StatePaused)
}

func TestTickRate(t *testing.T) {
	assert.InDelta(t, 0.05, tickInterval, 0.001)
}
