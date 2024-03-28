package raiden

type (
	Bucket interface {
		Name() string
		Public() bool
		AllowedMimeTypes() []string
		FileSizeLimit() int
		AvifAutoDetection() bool
	}

	BucketBase struct{}
)

func (b *BucketBase) Public() bool {
	return false
}

func (b *BucketBase) AllowedMimeTypes() []string {
	return nil
}

func (b *BucketBase) FileSizeLimit() *int {
	return nil
}

func (b *BucketBase) AvifAutoDetection() bool {
	return false
}
