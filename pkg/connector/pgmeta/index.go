package pgmeta

import (
	"fmt"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
)

func GetIndexes(cfg *raiden.Config, schema string) ([]objects.Index, error) {
	MetaLogger.Trace("start fetching indexes from meta")
	rs, err := ExecuteQuery[[]objects.Index](cfg.PgMetaUrl, sql.GenerateGetIndexQuery(schema), nil, DefaultAuthInterceptor(cfg.JwtToken), nil)
	if err != nil {
		err = fmt.Errorf("get indexes error : %s", err)
		return []objects.Index{}, err
	}
	MetaLogger.Trace("finish fetching policy by name from meta")
	return rs, nil
}
