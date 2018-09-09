package quarry

type stringSet map[string]struct{}

func newStringSet() stringSet {
	return make(stringSet)
}

func (s stringSet) Remove(val string) {
	delete(s, val)
}

func (s stringSet) Add(value string) {
	s[value] = struct{}{}
}

func (s stringSet) Contains(val string) bool {
	_, ok := s[val]
	return ok
}
