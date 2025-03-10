package meta

import (
	"fmt"

	"github.com/lib/pq"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
)

func GetUserByEmail(cfg *raiden.Config, email string) (result objects.User, err error) {
	MetaLogger.Trace("start fetching user by email from meta")
	qTemplate := sql.GenerateGetUserQuery(email)
	q := fmt.Sprintf(qTemplate, pq.QuoteLiteral(email))

	rs, err := ExecuteQuery[[]objects.User](getBaseUrl(cfg), q, nil, DefaultInterceptor(cfg), nil)
	if err != nil {
		err = fmt.Errorf("get email from error : %s", err)
		return
	}

	if len(rs) == 0 {
		err = fmt.Errorf("get email %s is not found", email)
		return
	}
	MetaLogger.Trace("finish fetching email by name from meta")
	return rs[0], nil
}

func UpdateUserRecoveryToken(cfg *raiden.Config, email string, token string) error {
	MetaLogger.Trace("start update user recovery token", "email", email)
	sql := sql.GenerateUpdateUserRecoveryTokenQuery(token, email)
	_, err := ExecuteQuery[any](getBaseUrl(cfg), sql, nil, DefaultInterceptor(cfg), nil)
	if err != nil {
		return fmt.Errorf("update user recovery token %s error : %s", email, err)
	}
	MetaLogger.Trace("finish update user recovery token", "user-id", email)
	return nil
}
