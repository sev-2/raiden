package raiden

type (
	BaseLibrary struct{}

	Library interface {
		IsLongRunning() bool
	}
)

func (b *BaseLibrary) IsLongRunning() bool {
	return false
}
