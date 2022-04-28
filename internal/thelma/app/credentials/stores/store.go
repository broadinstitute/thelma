package stores

// Store is a simple key-value interface for storing credentials in Thelma's root directory
type Store interface {
	Read(key string) ([]byte, error)
	Exists(key string) (bool, error)
	Write(key string, credential []byte) error
}
