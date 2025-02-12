package storages

import (
	"strings"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

func BuildGenerateStorageInput(storages []objects.Bucket, policies objects.Policies) []*generator.GenerateStorageInput {
	generateInputs := make([]*generator.GenerateStorageInput, 0)
	for i := range storages {
		s := storages[i]

		mapManagedPermissions := map[string]bool{
			supabase.GetPolicyName(objects.PolicyCommandInsert, strings.ToLower(supabase.RlsTypeStorage), s.Name): true,
			supabase.GetPolicyName(objects.PolicyCommandUpdate, strings.ToLower(supabase.RlsTypeStorage), s.Name): true,
			supabase.GetPolicyName(objects.PolicyCommandDelete, strings.ToLower(supabase.RlsTypeStorage), s.Name): true,
			supabase.GetPolicyName(objects.PolicyCommandSelect, strings.ToLower(supabase.RlsTypeStorage), s.Name): true,
		}

		p := policies.FilterByBucket(s)

		finalPolicy := objects.Policies{}
		for i := range p {
			pp := p[i]
			if _, exist := mapManagedPermissions[pp.Name]; exist {
				finalPolicy = append(finalPolicy, pp)
			}
		}

		generateInputs = append(generateInputs, &generator.GenerateStorageInput{
			Bucket:   s,
			Policies: finalPolicy,
		})
	}
	return generateInputs
}
