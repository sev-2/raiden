package generator_test

import (
	"encoding/json"
	"testing"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

const dummyPolicies = `
[
    {
        "id": 30023,
        "schema": "storage",
        "table": "objects",
        "table_id": 29647,
        "name": "enable select access for storage my-storage",
        "action": "PERMISSIVE",
        "roles": [
            "admin_scouter",
            "anon",
            "authenticated"
        ],
        "command": "SELECT",
        "definition": "(bucket_id = 'my-storage')",
        "check": null
    },
    {
        "id": 30022,
        "schema": "storage",
        "table": "objects",
        "table_id": 29647,
		"name": "enable update access for storage my-storage",
        "action": "PERMISSIVE",
        "roles": [
            "admin_scouter",
            "authenticated"
        ],
        "command": "UPDATE",
        "definition": "(bucket_id = 'my-storage')",
        "check": null
    },
    {
        "id": 30021,
        "schema": "storage",
        "table": "objects",
        "table_id": 29647,
		"name": "enable delete access for storage my-storage",
        "action": "PERMISSIVE",
        "roles": [
            "admin_scouter",
            "authenticated"
        ],
        "command": "DELETE",
        "definition": "(bucket_id = 'my-storage')",
        "check": null
    },
    {
        "id": 30020,
        "schema": "storage",
        "table": "objects",
        "table_id": 29647,
		"name": "enable insert access for storage my-storage",
        "action": "PERMISSIVE",
        "roles": [
            "admin_scouter",
            "authenticated"
        ],
        "command": "INSERT",
        "definition": "",
        "check": "(bucket_id = 'my-storage')"
    }
]
`

func TestBuildStorageRlsTag(t *testing.T) {
	var bucket = objects.Bucket{Name: "my-storage"}

	var policies objects.Policies
	err := json.Unmarshal([]byte(dummyPolicies), &policies)
	assert.NoError(t, err)

	storagePolicies := policies.FilterByBucket(bucket)
	rlsTag := generator.BuildRlsTag(storagePolicies, bucket.Name, supabase.RlsTypeStorage)
	expectedTag := `read:"admin_scouter,anon,authenticated" write:"admin_scouter,authenticated"`
	assert.Equal(t, expectedTag, rlsTag)
}
