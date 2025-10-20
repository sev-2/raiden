package storages

import (
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

func BuildGenerateStorageInput(storages []objects.Bucket, policies objects.Policies) []*generator.GenerateStorageInput {
	generateInputs := make([]*generator.GenerateStorageInput, 0)
	for i := range storages {
		s := storages[i]
		generateInputs = append(generateInputs, &generator.GenerateStorageInput{
			Bucket:   s,
			Policies: policies.FilterByBucket(s),
		})
	}
	return generateInputs
}
