package stores

func NewNoopStore() Store {
	return noopStore{}
}

type noopStore struct {
}

func (s noopStore) Read(_ string) ([]byte, error) {
	return nil, nil
}

func (s noopStore) Exists(_ string) (bool, error) {
	return false, nil
}

func (s noopStore) Write(_ string, _ []byte) error {
	return nil
}
