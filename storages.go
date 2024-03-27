package raiden

type (
	Storage interface {
		Id() string
		Name() string
		Public() bool
		AllowedMimeTypes() []string
		FileSizeLimit() *int
		AvifAutoDetection() bool
	}

	StorageBase struct{}
)

func (b *StorageBase) Public() bool {
	return false
}

func (b *StorageBase) AllowedMimeTypes() []string {
	return nil
}

func (b *StorageBase) FileSizeLimit() *int {
	return nil
}

func (b *StorageBase) AvifAutoDetection() bool {
	return false
}
