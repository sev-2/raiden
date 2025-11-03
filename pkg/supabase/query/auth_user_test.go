package query

import (
	"testing"

	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestUpdateUserData(t *testing.T) {
	query, err := UpdateUserData("00000000-0000-0000-0000-000000000000", objects.User{Role: "Admin"})
	assert.NoError(t, err)
	assert.Contains(t, query, "UPDATE auth.users SET role = 'Admin'")
	assert.Contains(t, query, "WHERE id = '00000000-0000-0000-0000-000000000000'")
}

func TestUpdateUserDataValidation(t *testing.T) {
	_, err := UpdateUserData("", objects.User{Role: "Admin"})
	assert.Error(t, err)
	_, err = UpdateUserData("user-id", objects.User{})
	assert.ErrorIs(t, err, ErrNoUserFieldsToUpdate)
}
