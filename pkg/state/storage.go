package state

import (
	"reflect"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type ExtractStorageItem struct {
	Storage           objects.Bucket
	ExtractedPolicies ExtractPolicyResult
}

type ExtractStorageResult struct {
	Existing []ExtractStorageItem
	New      []ExtractStorageItem
	Delete   []ExtractStorageItem
}

func ExtractStorage(storageStates []StorageState, appStorages []raiden.Bucket) (result ExtractStorageResult, err error) {
	mapStorageState := map[string]StorageState{}
	for i := range storageStates {
		s := storageStates[i]
		mapStorageState[s.Storage.Name] = s
	}

	for _, storage := range appStorages {
		state, isStateExist := mapStorageState[storage.Name()]
		if !isStateExist {
			si := BuildStorageFromApp(storage)
			result.New = append(result.New, si)
			continue
		}
		sr := BuildStorageFromState(state, storage)
		result.Existing = append(result.Existing, sr)

		delete(mapStorageState, storage.Name())
	}

	for _, state := range mapStorageState {
		result.Delete = append(result.Delete, ExtractStorageItem{
			Storage: state.Storage,
			ExtractedPolicies: ExtractPolicyResult{
				Delete: state.Policies,
			},
		})
	}

	return
}

func BindToSupabaseStorage(s *objects.Bucket, storage raiden.Bucket) {
	name := storage.Name()
	if name == "" {
		rv := reflect.TypeOf(storage)
		name = utils.ToSnakeCase(rv.Name())
	}

	s.Name = name
	s.Public = storage.Public()
	s.AllowedMimeTypes = storage.AllowedMimeTypes()
	s.AvifAutoDetection = storage.AvifAutoDetection()
	if storage.FileSizeLimit() > 0 {
		limit := storage.FileSizeLimit()
		s.FileSizeLimit = &limit
	} else {
		s.FileSizeLimit = nil
	}
}

func BuildStorageFromApp(storage raiden.Bucket) (si ExtractStorageItem) {
	BindToSupabaseStorage(&si.Storage, storage)

	if acl := getAcl(storage); acl != nil {
		policies, err := acl.BuildStoragePolicies(storage.Name())
		if err != nil {
			panic(err.Error())
		}
		si.ExtractedPolicies.New = append(si.ExtractedPolicies.New, policies...)
	}

	return
}

func BuildStorageFromState(ss StorageState, storage raiden.Bucket) (si ExtractStorageItem) {
	si.Storage = ss.Storage
	BindToSupabaseStorage(&si.Storage, storage)

	// map policies
	mapPolicies := make(map[string]objects.Policy)
	for i := range ss.Policies {
		p := ss.Policies[i]
		mapPolicies[p.Name] = p
	}

	if acl := getAcl(storage); acl != nil {
		policies, err := acl.BuildStoragePolicies(storage.Name())
		if err != nil {
			panic(err.Error())
		}
		for _, p := range policies {
			if sp, exist := mapPolicies[p.Name]; exist {
				sp.Roles = p.Roles
				sp.Check = p.Check
				sp.Definition = p.Definition
				si.ExtractedPolicies.Existing = append(si.ExtractedPolicies.Existing, sp)
				delete(mapPolicies, p.Name)
			} else {
				si.ExtractedPolicies.New = append(si.ExtractedPolicies.New, p)
			}
		}
	}

	for _, v := range mapPolicies {
		si.ExtractedPolicies.Delete = append(si.ExtractedPolicies.Delete, v)
	}
	return
}

func (er ExtractStorageResult) ToDeleteFlatMap() map[string]*objects.Bucket {
	mapData := make(map[string]*objects.Bucket)

	if len(er.Delete) > 0 {
		for i := range er.Delete {
			r := er.Delete[i]
			mapData[r.Storage.Name] = &r.Storage
		}
	}

	return mapData
}
