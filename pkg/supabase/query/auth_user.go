package query

import (
	"fmt"
	"strings"

	"github.com/sev-2/raiden/pkg/supabase/objects"
)

func UpdateUserData(userId string, data objects.User) string {
	setData := make([]string, 0)

	// TODO: handle all data change
	if data.Role != "" {
		setData = append(setData, fmt.Sprintf("role = '%s'", data.Role))
	}

	return fmt.Sprintf("UPDATE auth.users SET %s WHERE id = '%s';", strings.Join(setData, ","), userId)
}
