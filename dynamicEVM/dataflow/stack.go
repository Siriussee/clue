package dataflow

// stack tracks the data flow in addition to the EVM stack
type stack struct {
	s [][32]*DataSource
}

func newStack() *stack {
	return &stack{
		s: [][32]*DataSource{},
	}
}

func (s *stack) pop(n int) {
	s.s = s.s[:len(s.s)-n]
}

func (s *stack) push(value word) {
	s.s = append(s.s, value)
}

func (s *stack) dup(n int) {
	s.s = append(s.s, s.s[len(s.s)-n])
}

func (s *stack) swap(n int) {
	// swap the top and n-th element
	s.s[len(s.s)-1], s.s[len(s.s)-1-n] = s.s[len(s.s)-1-n], s.s[len(s.s)-1]
}

// func (s *stack) peek(i int) []*DataSource {
// 	return nil
// 	return s.s[len(s.s)-i-1][:]
// }

func (s *stack) peekNPush(n int) {
	if n <= 1 {
		panic("peekNPush: invalid n")
	}

	var sources []*DataSource
	for _, e := range s.s[len(s.s)-n : len(s.s)] {
		sources = append(sources, e[:]...)
	}

	source := mergeDataSources(sources...)

	var word [32]*DataSource
	for i := 0; i < 32; i++ {
		word[i] = source
	}

	s.s = append(s.s[:len(s.s)-n], word)
}
