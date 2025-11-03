package state

import "github.com/sev-2/raiden/pkg/supabase/objects"

type ExtractPolicyResult struct {
	Existing []objects.Policy
	New      []objects.Policy
	Delete   []objects.Policy
}

func (ep ExtractPolicyResult) ToDeleteFlatMap() map[string]*objects.Policy {
	mapData := make(map[string]*objects.Policy)

	if len(ep.Delete) > 0 {
		for i := range ep.Delete {
			r := ep.Delete[i]
			mapData[r.Name] = &r
		}
	}

	return mapData
}
