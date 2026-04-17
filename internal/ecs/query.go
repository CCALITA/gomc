package ecs

// Query2 iterates all entities that have both component A and B.
// It walks the smaller store and checks membership in the other for
// efficiency.
func Query2[A, B any](w *World, fn func(Entity, *A, *B)) {
	sa := GetStore[A](w)
	sb := GetStore[B](w)

	if sa.Len() <= sb.Len() {
		sa.Each(func(e Entity, a *A) {
			if b, ok := sb.Get(e); ok {
				fn(e, a, b)
			}
		})
	} else {
		sb.Each(func(e Entity, b *B) {
			if a, ok := sa.Get(e); ok {
				fn(e, a, b)
			}
		})
	}
}

// Query3 iterates all entities that have components A, B, and C.
// It walks the smallest store and checks the other two.
func Query3[A, B, C any](w *World, fn func(Entity, *A, *B, *C)) {
	sa := GetStore[A](w)
	sb := GetStore[B](w)
	sc := GetStore[C](w)

	// Find the smallest store to drive iteration.
	type iterChoice int
	const (
		iterA iterChoice = iota
		iterB
		iterC
	)

	minLen := sa.Len()
	choice := iterA
	if sb.Len() < minLen {
		minLen = sb.Len()
		choice = iterB
	}
	if sc.Len() < minLen {
		choice = iterC
	}

	switch choice {
	case iterA:
		sa.Each(func(e Entity, a *A) {
			if b, ok := sb.Get(e); ok {
				if c, ok := sc.Get(e); ok {
					fn(e, a, b, c)
				}
			}
		})
	case iterB:
		sb.Each(func(e Entity, b *B) {
			if a, ok := sa.Get(e); ok {
				if c, ok := sc.Get(e); ok {
					fn(e, a, b, c)
				}
			}
		})
	case iterC:
		sc.Each(func(e Entity, c *C) {
			if a, ok := sa.Get(e); ok {
				if b, ok := sb.Get(e); ok {
					fn(e, a, b, c)
				}
			}
		})
	}
}
