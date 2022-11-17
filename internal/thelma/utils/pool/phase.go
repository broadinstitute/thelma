package pool

// Phase is an enum type representing the execution phase of a work item
type Phase int64

const (
	Queued Phase = iota
	Running
	Success
	Error
)

func (p Phase) String() string {
	switch p {
	case Queued:
		return "queued"
	case Running:
		return "running"
	case Success:
		return "success"
	case Error:
		return "error"
	}
	return "unknown"
}
