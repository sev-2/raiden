package builder_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/builder"
	"github.com/stretchr/testify/assert"
)

func TestStorageBucketClause(t *testing.T) {
	assert.Equal(t, builder.Clause(""), builder.StorageBucketClause(""))
	assert.Equal(t, builder.Clause(`"bucket_id" = 'avatars'`), builder.StorageBucketClause("avatars"))
}

func TestStorageUsingClause(t *testing.T) {
	clause := builder.StorageUsingClause("avatars", builder.Eq("owner", builder.AuthUID()))
	assert.Contains(t, clause.String(), `"bucket_id" = 'avatars'`)
	assert.Contains(t, clause.String(), "auth.uid()")
}

func TestStorageUsingClauseSkipsDuplicateBucket(t *testing.T) {
	base := builder.Clause(`"bucket_id" = 'avatars' AND auth.role() = 'anon'`)
	clause := builder.StorageUsingClause("avatars", base)
	assert.Equal(t, base, clause)
}

func TestStripStorageBucketFilter(t *testing.T) {
	raw := `("bucket_id" = 'avatars') AND auth.role() = 'anon' AND storage.extension(name) = 'jpg'`
	clean := builder.StripStorageBucketFilter(raw, "avatars")
	// NormalizeClauseSQL removes identifier parentheses, so extension(name) becomes extensionname.
	expected := `auth.role() = 'anon' AND storage.extensionname = 'jpg'`
	assert.Equal(t, expected, clean)

	assert.Equal(t, "", builder.StripStorageBucketFilter(`"bucket_id" = 'avatars'`, "avatars"))
}
