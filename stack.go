package typegen

type Stack[S []T, T comparable] []T

func (s *Stack[S, T]) Push(v T) {
	*s = append(*s, v)
}

func (s *Stack[S, T]) Has(v T) (has bool) {
	for _, e := range *s {
		if e == v {
			return true
		}
	}
	return false
}

func (s *Stack[S, T]) Pop() T {
	res := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return res
}
func (s *Stack[S, T]) Drop() {
	*s = (*s)[:len(*s)-1]
}
func (s *Stack[S, T]) Empty() bool {
	return len(*s) == 0
}
func (s *Stack[S, T]) Depth() int {
	return len(*s)
}

func (s *Stack[S, T]) Top() T {
	return (*s)[len(*s)-1]
}
