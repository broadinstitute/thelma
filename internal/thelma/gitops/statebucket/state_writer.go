package statebucket

type transformFn func(input StateFile) (output StateFile, err error)

type writer interface {
	read() (StateFile, error)
	write(StateFile) error
	update(transformFn) error
}
