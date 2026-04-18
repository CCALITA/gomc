package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateManager_InitialState(t *testing.T) {
	sm := NewStateManager()
	assert.Equal(t, GameStateMainMenu, sm.CurrentState())
}

func TestStateManager_SetStateChangesState(t *testing.T) {
	sm := NewStateManager()

	sm.SetState(GameStatePlaying)
	assert.Equal(t, GameStatePlaying, sm.CurrentState())

	sm.SetState(GameStatePaused)
	assert.Equal(t, GameStatePaused, sm.CurrentState())

	sm.SetState(GameStateDead)
	assert.Equal(t, GameStateDead, sm.CurrentState())
}

func TestStateManager_SetStateSameStateIsNoOp(t *testing.T) {
	sm := NewStateManager()
	enterCalled := false
	exitCalled := false

	sm.OnEnter(GameStateMainMenu, func() { enterCalled = true })
	sm.OnExit(GameStateMainMenu, func() { exitCalled = true })

	sm.SetState(GameStateMainMenu)

	assert.False(t, enterCalled, "OnEnter should not fire when setting same state")
	assert.False(t, exitCalled, "OnExit should not fire when setting same state")
}

func TestStateManager_OnEnterCallbackFires(t *testing.T) {
	sm := NewStateManager()
	entered := false

	sm.OnEnter(GameStatePlaying, func() { entered = true })
	sm.SetState(GameStatePlaying)

	assert.True(t, entered, "OnEnter callback should fire on transition into state")
}

func TestStateManager_OnExitCallbackFires(t *testing.T) {
	sm := NewStateManager()
	exited := false

	sm.OnExit(GameStateMainMenu, func() { exited = true })
	sm.SetState(GameStatePlaying)

	assert.True(t, exited, "OnExit callback should fire on transition out of state")
}

func TestStateManager_ExitFiresBeforeEnter(t *testing.T) {
	sm := NewStateManager()
	var log []string

	sm.OnExit(GameStateMainMenu, func() { log = append(log, "exit-menu") })
	sm.OnEnter(GameStatePlaying, func() { log = append(log, "enter-playing") })

	sm.SetState(GameStatePlaying)

	assert.Equal(t, []string{"exit-menu", "enter-playing"}, log)
}

func TestStateManager_MultipleTransitionsInSequence(t *testing.T) {
	sm := NewStateManager()
	var log []string

	sm.OnExit(GameStateMainMenu, func() { log = append(log, "exit-menu") })
	sm.OnEnter(GameStatePlaying, func() { log = append(log, "enter-playing") })
	sm.OnExit(GameStatePlaying, func() { log = append(log, "exit-playing") })
	sm.OnEnter(GameStatePaused, func() { log = append(log, "enter-paused") })
	sm.OnExit(GameStatePaused, func() { log = append(log, "exit-paused") })
	sm.OnEnter(GameStateDead, func() { log = append(log, "enter-dead") })

	sm.SetState(GameStatePlaying)
	sm.SetState(GameStatePaused)
	sm.SetState(GameStateDead)

	expected := []string{
		"exit-menu",
		"enter-playing",
		"exit-playing",
		"enter-paused",
		"exit-paused",
		"enter-dead",
	}
	assert.Equal(t, expected, log)
}

func TestStateManager_UnrelatedCallbackDoesNotFire(t *testing.T) {
	sm := NewStateManager()
	pausedEntered := false
	deadEntered := false

	sm.OnEnter(GameStatePaused, func() { pausedEntered = true })
	sm.OnEnter(GameStateDead, func() { deadEntered = true })

	// Transition MainMenu -> Playing; neither Paused nor Dead callbacks should fire.
	sm.SetState(GameStatePlaying)

	assert.False(t, pausedEntered, "OnEnter for Paused should not fire during MainMenu->Playing")
	assert.False(t, deadEntered, "OnEnter for Dead should not fire during MainMenu->Playing")
}

func TestStateManager_SetStateWithNoCallbacksRegistered(t *testing.T) {
	sm := NewStateManager()

	// No callbacks registered at all; should not panic.
	sm.SetState(GameStatePlaying)
	sm.SetState(GameStatePaused)
	sm.SetState(GameStateDead)
	sm.SetState(GameStateMainMenu)

	assert.Equal(t, GameStateMainMenu, sm.CurrentState())
}

func TestStateManager_OnEnterCanBeOverwritten(t *testing.T) {
	sm := NewStateManager()
	firstCalled := false
	secondCalled := false

	sm.OnEnter(GameStatePlaying, func() { firstCalled = true })
	sm.OnEnter(GameStatePlaying, func() { secondCalled = true })

	sm.SetState(GameStatePlaying)

	assert.False(t, firstCalled, "first OnEnter should not fire after being overwritten")
	assert.True(t, secondCalled, "second OnEnter should fire after overwriting the first")
}

func TestStateManager_OnExitCanBeOverwritten(t *testing.T) {
	sm := NewStateManager()
	firstCalled := false
	secondCalled := false

	sm.OnExit(GameStateMainMenu, func() { firstCalled = true })
	sm.OnExit(GameStateMainMenu, func() { secondCalled = true })

	sm.SetState(GameStatePlaying)

	assert.False(t, firstCalled, "first OnExit should not fire after being overwritten")
	assert.True(t, secondCalled, "second OnExit should fire after overwriting the first")
}

func TestStateManager_OnEnterDoesNotAffectOnExit(t *testing.T) {
	sm := NewStateManager()
	exitCalled := false

	sm.OnExit(GameStateMainMenu, func() { exitCalled = true })
	// Setting OnEnter for the same state should not remove the OnExit.
	sm.OnEnter(GameStateMainMenu, func() {})

	sm.SetState(GameStatePlaying)

	assert.True(t, exitCalled, "OnExit should still fire after OnEnter was set for the same state")
}

func TestStateManager_RoundTripTransition(t *testing.T) {
	sm := NewStateManager()
	enterCount := 0

	sm.OnEnter(GameStatePlaying, func() { enterCount++ })

	sm.SetState(GameStatePlaying)
	sm.SetState(GameStatePaused)
	sm.SetState(GameStatePlaying)

	assert.Equal(t, 2, enterCount, "OnEnter should fire each time the state is entered")
	assert.Equal(t, GameStatePlaying, sm.CurrentState())
}
