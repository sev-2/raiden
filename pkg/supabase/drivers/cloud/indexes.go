package cloud

import (
	"fmt"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
)

func GetIndexes(cfg *raiden.Config, schema string) ([]objects.Index, error) {
	CloudLogger.Trace("start fetching index from supabase")
	rs, err := ExecuteQuery[[]objects.Index](cfg.SupabaseApiUrl, cfg.ProjectId, sql.GenerateGetIndexQuery(schema), DefaultAuthInterceptor(cfg.AccessToken), getRoleResponseInterceptor)
	if err != nil {
		err = fmt.Errorf("get index error : %s", err)
	}
	CloudLogger.Trace("finish fetching index from supabase")
	return rs, err
}
