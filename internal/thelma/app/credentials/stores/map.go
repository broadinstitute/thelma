package stores

// NewMapStore returns a new credential store that caches credentials in memory instead of on disk.
func NewMapStore() Store {
	return mapStore{
		_map: make(map[string][]byte),
	}
}

type mapStore struct {
	_map map[string][]byte
}

func (s mapStore) Read(key string) ([]byte, error) {
	return s._map[key], nil
}

func (s mapStore) Exists(key string) (bool, error) {
	_, exists := s._map[key]
	return exists, nil
}

func (s mapStore) Write(key string, token []byte) error {
	s._map[key] = token
	return nil
}

func (s mapStore) Remove(key string) error {
	delete(s._map, key)
	return nil
}
