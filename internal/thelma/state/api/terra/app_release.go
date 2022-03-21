package terra

type AppRelease interface {
	AppVersion() string
	Environment() Environment
	Release
}
