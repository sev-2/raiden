package query

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

var ErrNoUserFieldsToUpdate = errors.New("no user fields to update")

func UpdateUserData(userID string, data objects.User) (string, error) {
	if strings.TrimSpace(userID) == "" {
		return "", errors.New("user id is required")
	}

	setData := make([]string, 0)

	// TODO: handle all data change
	if data.Role != "" {
		setData = append(setData, fmt.Sprintf("role = %s", pq.QuoteLiteral(data.Role)))
	}

	if len(setData) == 0 {
		return "", ErrNoUserFieldsToUpdate
	}

	return fmt.Sprintf("UPDATE auth.users SET %s WHERE id = %s;", strings.Join(setData, ","), pq.QuoteLiteral(userID)), nil
}
