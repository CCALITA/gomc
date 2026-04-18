package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateManager_InitialState(t *testing.T) {
	sm := NewStateManager()
	assert.Equal(t, StateMainMenu, sm.CurrentState())
}

func TestStateManager_SetStateChangesState(t *testing.T) {
	sm := NewStateManager()

	sm.SetState(StatePlaying)
	assert.Equal(t, StatePlaying, sm.CurrentState())

	sm.SetState(StatePaused)
	assert.Equal(t, StatePaused, sm.CurrentState())

	sm.SetState(StateDead)
	assert.Equal(t, StateDead, sm.CurrentState())
}

func TestStateManager_SetStateSameStateIsNoOp(t *testing.T) {
	sm := NewStateManager()
	enterCalled := false
	exitCalled := false

	sm.OnEnter(StateMainMenu, func() { enterCalled = true })
	sm.OnExit(StateMainMenu, func() { exitCalled = true })

	sm.SetState(StateMainMenu)

	assert.False(t, enterCalled, "OnEnter should not fire when setting same state")
	assert.False(t, exitCalled, "OnExit should not fire when setting same state")
}

func TestStateManager_OnEnterCallbackFires(t *testing.T) {
	sm := NewStateManager()
	entered := false

	sm.OnEnter(StatePlaying, func() { entered = true })
	sm.SetState(StatePlaying)

	assert.True(t, entered, "OnEnter callback should fire on transition into state")
}

func TestStateManager_OnExitCallbackFires(t *testing.T) {
	sm := NewStateManager()
	exited := false

	sm.OnExit(StateMainMenu, func() { exited = true })
	sm.SetState(StatePlaying)

	assert.True(t, exited, "OnExit callback should fire on transition out of state")
}

func TestStateManager_ExitFiresBeforeEnter(t *testing.T) {
	sm := NewStateManager()
	var log []string

	sm.OnExit(StateMainMenu, func() { log = append(log, "exit-menu") })
	sm.OnEnter(StatePlaying, func() { log = append(log, "enter-playing") })

	sm.SetState(StatePlaying)

	assert.Equal(t, []string{"exit-menu", "enter-playing"}, log)
}

func TestStateManager_MultipleTransitionsInSequence(t *testing.T) {
	sm := NewStateManager()
	var log []string

	sm.OnExit(StateMainMenu, func() { log = append(log, "exit-menu") })
	sm.OnEnter(StatePlaying, func() { log = append(log, "enter-playing") })
	sm.OnExit(StatePlaying, func() { log = append(log, "exit-playing") })
	sm.OnEnter(StatePaused, func() { log = append(log, "enter-paused") })
	sm.OnExit(StatePaused, func() { log = append(log, "exit-paused") })
	sm.OnEnter(StateDead, func() { log = append(log, "enter-dead") })

	sm.SetState(StatePlaying)
	sm.SetState(StatePaused)
	sm.SetState(StateDead)

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

	sm.OnEnter(StatePaused, func() { pausedEntered = true })
	sm.OnEnter(StateDead, func() { deadEntered = true })

	// Transition MainMenu -> Playing; neither Paused nor Dead callbacks should fire.
	sm.SetState(StatePlaying)

	assert.False(t, pausedEntered, "OnEnter for Paused should not fire during MainMenu->Playing")
	assert.False(t, deadEntered, "OnEnter for Dead should not fire during MainMenu->Playing")
}

func TestStateManager_SetStateWithNoCallbacksRegistered(t *testing.T) {
	sm := NewStateManager()

	// No callbacks registered at all; should not panic.
	sm.SetState(StatePlaying)
	sm.SetState(StatePaused)
	sm.SetState(StateDead)
	sm.SetState(StateMainMenu)

	assert.Equal(t, StateMainMenu, sm.CurrentState())
}

func TestStateManager_OnEnterCanBeOverwritten(t *testing.T) {
	sm := NewStateManager()
	firstCalled := false
	secondCalled := false

	sm.OnEnter(StatePlaying, func() { firstCalled = true })
	sm.OnEnter(StatePlaying, func() { secondCalled = true })

	sm.SetState(StatePlaying)

	assert.False(t, firstCalled, "first OnEnter should not fire after being overwritten")
	assert.True(t, secondCalled, "second OnEnter should fire after overwriting the first")
}

func TestStateManager_OnExitCanBeOverwritten(t *testing.T) {
	sm := NewStateManager()
	firstCalled := false
	secondCalled := false

	sm.OnExit(StateMainMenu, func() { firstCalled = true })
	sm.OnExit(StateMainMenu, func() { secondCalled = true })

	sm.SetState(StatePlaying)

	assert.False(t, firstCalled, "first OnExit should not fire after being overwritten")
	assert.True(t, secondCalled, "second OnExit should fire after overwriting the first")
}

func TestStateManager_OnEnterDoesNotAffectOnExit(t *testing.T) {
	sm := NewStateManager()
	exitCalled := false

	sm.OnExit(StateMainMenu, func() { exitCalled = true })
	// Setting OnEnter for the same state should not remove the OnExit.
	sm.OnEnter(StateMainMenu, func() {})

	sm.SetState(StatePlaying)

	assert.True(t, exitCalled, "OnExit should still fire after OnEnter was set for the same state")
}

func TestStateManager_RoundTripTransition(t *testing.T) {
	sm := NewStateManager()
	enterCount := 0

	sm.OnEnter(StatePlaying, func() { enterCount++ })

	sm.SetState(StatePlaying)
	sm.SetState(StatePaused)
	sm.SetState(StatePlaying)

	assert.Equal(t, 2, enterCount, "OnEnter should fire each time the state is entered")
	assert.Equal(t, StatePlaying, sm.CurrentState())
}
