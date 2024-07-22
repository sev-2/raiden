package storages_test

import (
	"encoding/json"
	"testing"

	"github.com/sev-2/raiden/pkg/resource/storages"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestBuildGenerateStorageInputs(t *testing.T) {
	jsonStrData := `[{"name":"some-bucket"},{"name":"another-bucket"}]`

	var sourceStorages []objects.Bucket
	err := json.Unmarshal([]byte(jsonStrData), &sourceStorages)
	assert.NoError(t, err)

	rs := storages.BuildGenerateStorageInput(sourceStorages, nil)

	for _, r := range rs {
		assert.NotNil(t, r.Bucket)
	}
}
