package typegen

type Stack[T comparable] []T

func (s *Stack[T]) Push(v T) {
	*s = append(*s, v)
}

func (s *Stack[T]) Has(v T) (has bool) {
	for _, e := range *s {
		if e == v {
			return true
		}
	}
	return false
}

func (s *Stack[T]) Pop() T {
	res := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return res
}
func (s *Stack[T]) Drop() {
	*s = (*s)[:len(*s)-1]
}
func (s *Stack[T]) Empty() bool {
	return len(*s) == 0
}
func (s *Stack[T]) Depth() int {
	return len(*s)
}

func (s *Stack[T]) Top() T {
	return (*s)[len(*s)-1]
}
