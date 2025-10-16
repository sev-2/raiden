package meta

import (
	"errors"
	"fmt"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query"
)

func UpdateUserData(cfg *raiden.Config, userId string, t objects.User) error {
	MetaLogger.Trace("start update user data", "id", userId)
	updateSql, err := query.UpdateUserData(userId, t)
	if err != nil {
		if errors.Is(err, query.ErrNoUserFieldsToUpdate) {
			MetaLogger.Trace("skip update user data, no fields to update", "id", userId)
			return nil
		}
		return err
	}
	_, err = ExecuteQuery[any](getBaseUrl(cfg), updateSql, nil, DefaultInterceptor(cfg), nil)
	if err != nil {
		return fmt.Errorf("update user data with id '%s' error : %s", userId, err)
	}
	MetaLogger.Trace("finish update user data", "id", userId)
	return nil
}
