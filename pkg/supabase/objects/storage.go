package objects

type Bucket struct {
	ID                string   `json:"id,omitempty"`
	Name              string   `json:"name,omitempty"`
	Owner             *string  `json:"owner,omitempty"`
	Public            bool     `json:"public"`
	AvifAutoDetection bool     `json:"avif_autodetection,omitempty"`
	FileSizeLimit     *int     `json:"file_size_limit"`
	AllowedMimeTypes  []string `json:"allowed_mime_types"`
	OwnerID           *string  `json:"owner_id,omitempty"`
}

type UpdateBucketType string

const (
	UpdateBucketIsPublic         UpdateBucketType = "is_public"
	UpdateBucketFileSizeLimit    UpdateBucketType = "file_size_limit"
	UpdateBucketAllowedMimeTypes UpdateBucketType = "allowed_mime_types"
)

type UpdateBucketParam struct {
	OldData     Bucket
	ChangeItems []UpdateBucketType
}
