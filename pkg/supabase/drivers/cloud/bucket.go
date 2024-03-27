package cloud

import (
	"fmt"

	"github.com/lib/pq"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
)

func GetStorages(cfg *raiden.Config) ([]objects.Storage, error) {
	rs, err := ExecuteQuery[[]objects.Storage](
		cfg.SupabaseApiUrl, cfg.ProjectId, sql.GetBucketsQuery, DefaultAuthInterceptor(cfg.AccessToken), nil,
	)
	if err != nil {
		err = fmt.Errorf("get storage error : %s", err)
	}
	return rs, err
}

func GetStorageByName(cfg *raiden.Config, name string) (result objects.Storage, err error) {
	qTemplate := sql.GetRolesQuery + " where name = %s limit 1"
	q := fmt.Sprintf(qTemplate, pq.QuoteLiteral(name))

	rs, err := ExecuteQuery[[]objects.Storage](cfg.SupabaseApiUrl, cfg.ProjectId, q, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		err = fmt.Errorf("get bucket error : %s", err)
		return
	}

	if len(rs) == 0 {
		err = fmt.Errorf("get bucket %s is not found", name)
		return
	}

	return rs[0], nil
}

func CreateStorage(cfg *raiden.Config, bucket objects.Storage) (objects.Storage, error) {
	sql := query.BuildBucketQuery(bucket)
	// Execute SQL Query
	logger.Debug("Create Bucket - execute : ", sql)
	_, err := ExecuteQuery[any](cfg.SupabaseApiUrl, cfg.ProjectId, sql, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		return objects.Storage{}, fmt.Errorf("create new bucket %s error : %s", bucket.Name, err)
	}
	return GetStorageByName(cfg, bucket.Name)
}
