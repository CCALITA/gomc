package game

type State int

const (
	StateMainMenu State = iota
	StatePlaying
	StatePaused
	StateDead
)

type stateCallbacks struct {
	onEnter func()
	onExit  func()
}

type StateManager struct {
	current   State
	callbacks map[State]stateCallbacks
}

func NewStateManager() *StateManager {
	return &StateManager{
		current:   StateMainMenu,
		callbacks: make(map[State]stateCallbacks),
	}
}

func (sm *StateManager) CurrentState() State {
	return sm.current
}

func (sm *StateManager) SetState(s State) {
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

func (sm *StateManager) OnEnter(s State, fn func()) {
	cb := sm.callbacks[s]
	cb.onEnter = fn
	sm.callbacks[s] = cb
}

func (sm *StateManager) OnExit(s State, fn func()) {
	cb := sm.callbacks[s]
	cb.onExit = fn
	sm.callbacks[s] = cb
}
