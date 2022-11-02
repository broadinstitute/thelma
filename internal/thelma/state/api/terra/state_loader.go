package terra

type StateLoader interface {
	Load() (State, error)
	Reload() (State, error)
}
