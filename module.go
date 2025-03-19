package raiden

type (
	BaseModule struct{}

	Module interface {
		Routes() []*Route
		Libs() []func(config *Config) any
	}
)
