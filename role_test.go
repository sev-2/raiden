package raiden_test

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/postgres/roles"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalRoleMetadataTag(t *testing.T) {
	metaTag := `config:"key1:value1;key2:value2" connectionLimit:"60" inheritRole:"true" isReplicationRole:"false" isSuperuser:"true"`
	metaTagInstance := raiden.UnmarshalRoleMetadataTag(metaTag)

	assert.Equal(t, 2, len(metaTagInstance.Config))
	assert.Equal(t, "value1", metaTagInstance.Config["key1"])
	assert.Equal(t, "value2", metaTagInstance.Config["key2"])
	assert.Equal(t, 60, metaTagInstance.ConnectionLimit)
	assert.True(t, metaTagInstance.InheritRole)
	assert.False(t, metaTagInstance.IsReplicationRole)
	assert.True(t, metaTagInstance.IsSuperuser)
}

func TestMarshalRoleMetadataTag(t *testing.T) {
	meta := &raiden.RoleMetadataTag{
		Config:            map[string]any{},
		ConnectionLimit:   60,
		InheritRole:       true,
		IsReplicationRole: true,
		IsSuperuser:       true,
	}

	expectedMetaTagString := `config:"" connectionLimit:"60" inheritRole:"true" isReplicationRole:"true" isSuperuser:"true"`
	metaTagString := raiden.MarshalRoleMetadataTag(meta)
	assert.Equal(t, expectedMetaTagString, metaTagString)
}

func TestUnmarshalRole(t *testing.T) {
	// `config:"statement_timeout:3s" connectionLimit:"60" inheritRole:"true" isReplicationRole:"false" isSuperuser:"false"`
	// `canBypassRls:"false" canCreateDb:"false" canCreateRole:"false" canLogin:"false"`
	role := &roles.Anon{}

	anonRole, err := raiden.UnmarshalRole(role)
	assert.NoError(t, err)

	assert.Equal(t, "anon", anonRole.Name)
	assert.Equal(t, 60, anonRole.ConnectionLimit)
	assert.True(t, anonRole.InheritRole)
	assert.False(t, anonRole.IsReplicationRole)
	assert.False(t, anonRole.IsSuperuser)
	assert.False(t, anonRole.CanBypassRLS)
	assert.False(t, anonRole.CanCreateDB)
	assert.False(t, anonRole.CanCreateRole)
	assert.False(t, anonRole.CanLogin)
}
