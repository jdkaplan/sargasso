package gsync

type Set[K comparable] Map[K, struct{}]

func NewSet[K comparable]() *Set[K] {
	return (*Set[K])(NewMap[K, struct{}]())
}

func (s *Set[K]) Add(key K) {
	s.m.Store(key, struct{}{})
}

func (s *Set[K]) Has(key K) bool {
	_, ok := s.m.Load(key)
	return ok
}

func (s *Set[K]) Del(key K) {
	s.m.Delete(key)
}

func (s *Set[K]) Values() []K {
	return (*Map[K, struct{}])(s).Keys()
}
