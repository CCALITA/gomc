package game

type GameState int

const (
	GameStateMainMenu GameState = iota
	GameStatePlaying
	GameStatePaused
	GameStateInventory
	GameStateDead
)

type stateCallbacks struct {
	onEnter func()
	onExit  func()
}

type StateManager struct {
	current   GameState
	callbacks map[GameState]stateCallbacks
}

func NewStateManager() *StateManager {
	return &StateManager{
		current:   GameStateMainMenu,
		callbacks: make(map[GameState]stateCallbacks),
	}
}

func (sm *StateManager) CurrentState() GameState {
	return sm.current
}

func (sm *StateManager) SetState(s GameState) {
	if sm.current == s {
		return
	}
	if cb, ok := sm.callbacks[sm.current]; ok && cb.onExit != nil {
		cb.onExit()
	}
	sm.current = s
	if cb, ok := sm.callbacks[s]; ok && cb.onEnter != nil {
		cb.onEnter()
	}
}

func (sm *StateManager) OnEnter(s GameState, fn func()) {
	cb := sm.callbacks[s]
	cb.onEnter = fn
	sm.callbacks[s] = cb
}

func (sm *StateManager) OnExit(s GameState, fn func()) {
	cb := sm.callbacks[s]
	cb.onExit = fn
	sm.callbacks[s] = cb
}
