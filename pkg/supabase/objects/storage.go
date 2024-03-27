package objects

type Storage struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Owner             *string  `json:"owner"`
	Public            bool     `json:"public"`
	AvifAutoDetection bool     `json:"avif_autodetection"`
	FileSizeLimit     *int     `json:"file_size_limit"`
	AllowedMimeTypes  []string `json:"allowed_mime_types"`
	OwnerID           *string  `json:"owner_id"`
}

type UpdateStorageType string

const (
	UpdateStorageName              UpdateStorageType = "name"
	UpdateStorageIsPublic          UpdateStorageType = "is_public"
	UpdateStorageAvifAutoDetection UpdateStorageType = "is_avif_auto_detection"
	UpdateStorageFileSizeLimit     UpdateStorageType = "file_size_limit"
	UpdateStorageAllowedMimeTypes  UpdateStorageType = "allowed_mime_types"
)

type UpdateStorageParam struct {
	OldData     Storage
	ChangeItems []UpdateStorageType
}
