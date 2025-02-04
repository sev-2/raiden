package cloud

import (
	"fmt"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
)

func GetTypes(cfg *raiden.Config, includedSchemas []string) ([]objects.Type, error) {
	CloudLogger.Trace("start fetching types from supabase")
	q := sql.GenerateTypeQuery(includedSchemas)
	rs, err := ExecuteQuery[[]objects.Type](cfg.SupabaseApiUrl, cfg.ProjectId, q, DefaultAuthInterceptor(cfg.AccessToken), nil)
	if err != nil {
		err = fmt.Errorf("get types error : %s", err)
	}
	CloudLogger.Trace("finish fetching types from supabase")
	return rs, err
}
