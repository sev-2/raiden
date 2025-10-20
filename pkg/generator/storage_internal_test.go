package generator

import (
	"io"
	"testing"

	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/require"
)

func TestGenerateStorageDataBinding(t *testing.T) {
	fileLimit := 512
	input := &GenerateStorageInput{
		Bucket: objects.Bucket{
			Name:              "user_avatars",
			Public:            true,
			FileSizeLimit:     &fileLimit,
			AvifAutoDetection: true,
			AllowedMimeTypes:  []string{"image/png", "image/jpeg"},
		},
		Policies: objects.Policies{},
	}

	var captured GenerateInput
	fakeFn := func(in GenerateInput, _ io.Writer) error {
		captured = in
		return nil
	}

	err := GenerateStorage("/tmp", "github.com/example/project", input, nil, nil, fakeFn)
	require.NoError(t, err)

	data, ok := captured.BindData.(GenerateStoragesData)
	require.True(t, ok)
	require.Equal(t, "storages", data.Package)
	require.Equal(t, "UserAvatars", data.StructName)
	require.Equal(t, "user_avatars", data.Name)
	require.True(t, data.Public)
	require.Equal(t, fileLimit, data.FileSizeLimit)
	require.True(t, data.AvifAutoDetection)
	require.Contains(t, data.AllowedMimeTypes, "image/png")
	require.Contains(t, data.AllowedMimeTypes, "image/jpeg")
	require.Contains(t, data.Imports, "\"github.com/sev-2/raiden\"")
}
