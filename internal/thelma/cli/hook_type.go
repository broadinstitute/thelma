package cli

// hookType is an enum type representing the different types of run hooks supported by ThelmaCLI
type hookType int

const (
	// setFlagsFromEnvironment isn't a "hook" per se, but it's a phase within an execution that can fail and it doesn't
	// participate in the same ordering as any of the real hooks. We include it here to make errors more informative.
	setFlagsFromEnvironment hookType = iota
	preRunHook
	runHook
	postRunHook
)

func (h hookType) String() string {
	switch h {
	case setFlagsFromEnvironment:
		return "set-flags-from-environment"
	case preRunHook:
		return "pre-run"
	case runHook:
		return "run"
	case postRunHook:
		return "post-run"
	}
	return "unknown"
}
