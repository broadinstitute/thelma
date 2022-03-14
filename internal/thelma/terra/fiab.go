package terra

// Fiab (DEPRECATED) represents a legacy Fiab ("firecloud-in-a-box") environment
type Fiab interface {
	// IP returns the public IP address for the Fiab
	IP() string
	// Name returns the name of the Fiab
	Name() string
}

// NewFiab constructor for a new Fiab
func NewFiab(name string, ip string) Fiab {
	return &fiab{
		name: name,
		ip:   ip,
	}
}

// DEPRECATED implements the Fiab interface
type fiab struct {
	name string
	ip   string
}

func (f *fiab) Name() string {
	return f.name
}

func (f *fiab) IP() string {
	return f.ip
}
