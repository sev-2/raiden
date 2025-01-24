package storages

import (
	"github.com/sev-2/raiden"
)

type MyStorage struct {
	raiden.BucketBase

	// Access control
	Acl string `json:"-" read:"admin_scouter,authenticated" write:"admin_scouter,authenticated"`
}

func (r *MyStorage) Name() string {
	return "my-storage"
}
func (r *MyStorage) Public() bool {
	return true
}
