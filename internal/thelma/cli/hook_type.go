package cli

// hookType is an enum type representing the different types of run hooks supported by ThelmaCLI
type hookType int

const (
	preRunHook hookType = iota
	runHook
	postRunHook
)

func (h hookType) String() string {
	switch h {
	case preRunHook:
		return "pre-run"
	case runHook:
		return "run"
	case postRunHook:
		return "post-run"
	}
	return "unknown"
}
