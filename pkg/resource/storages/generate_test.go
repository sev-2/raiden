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
func TestBuildGenerateStorageInputs_WithPolicies(t *testing.T) {
	storagesList := []objects.Bucket{
		{
			ID:   "test-bucket",
			Name: "test-bucket",
			Public: true,
		},
	}
	
	// Based on the function, managed permissions are built with specific patterns:
	// supabase.GetPolicyName(objects.PolicyCommandInsert, strings.ToLower(supabase.RlsTypeStorage), s.Name)
	// Which creates names like: "enable insert access for storage test-bucket"
	// FilterByBucket looks for policies where Schema="storage" and Definition or Check contains bucket name
	insertPolicyName := "enable insert access for storage test-bucket"
	updatePolicyName := "enable update access for storage test-bucket"
	
	policies := objects.Policies{
		{
			Name: insertPolicyName,  // Managed permission for insert
			Schema: "storage",
			Table: "test-bucket",
			Definition: "test-bucket", // Contains bucket name to pass FilterByBucket
		},
		{
			Name: updatePolicyName,  // Managed permission for update
			Schema: "storage", 
			Table: "test-bucket",
			Definition: "test-bucket", // Contains bucket name to pass FilterByBucket
		},
		{
			Name: "non-managed-policy",  // Not a managed permission
			Schema: "storage",
			Table: "test-bucket",
			Definition: "test-bucket", // But this one will be filtered out
		},
	}
	
	result := storages.BuildGenerateStorageInput(storagesList, policies)
	assert.Len(t, result, 1)
	assert.Len(t, result[0].Policies, 2) // Should have 2 matching managed policies (insert + update)
}

func TestBuildGenerateStorageInputs_Empty(t *testing.T) {
	result := storages.BuildGenerateStorageInput([]objects.Bucket{}, nil)
	assert.Empty(t, result)
}
