package query

import (
	"fmt"

	"github.com/sev-2/raiden/pkg/supabase/objects"
)

func BuildBucketQuery(bucket objects.Storage) string {
	isPublicStr := "true"
	if !bucket.Public {
		isPublicStr = "false"
	}

	return fmt.Sprintf(`insert into storage.buckets (id, name, public) values ('%s', '%s', %s);`,
		bucket.Name, bucket.Name, isPublicStr,
	)
}
