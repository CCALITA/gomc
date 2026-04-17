package ecs

// System is the interface every game system must implement.
type System interface {
	Update(w *World, dt float64)
}

// Scheduler holds an ordered list of systems and runs them sequentially.
type Scheduler struct {
	systems []System
}

// NewScheduler returns a ready-to-use Scheduler.
func NewScheduler() *Scheduler {
	return &Scheduler{}
}

// Add appends a system to the execution list.
func (s *Scheduler) Add(sys System) {
	s.systems = append(s.systems, sys)
}

// Update runs every registered system in insertion order.
func (s *Scheduler) Update(w *World, dt float64) {
	for _, sys := range s.systems {
		sys.Update(w, dt)
	}
}
