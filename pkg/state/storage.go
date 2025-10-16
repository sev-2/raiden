package state

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase"
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

	// build acl
	storageType := reflect.TypeOf(storage)
	if storageType.Kind() == reflect.Ptr {
		storageType = storageType.Elem()
	}

	aclField, isExist := storageType.FieldByName("Acl")
	if isExist {
		si.ExtractedPolicies.New = getStoragePolicies(&aclField, &si)
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

	// build acl
	storageType := reflect.TypeOf(storage)
	if storageType.Kind() == reflect.Ptr {
		storageType = storageType.Elem()
	}

	aclField, isExist := storageType.FieldByName("Acl")
	if isExist {
		policies := getStoragePolicies(&aclField, &si)
		for ip := range policies {
			p := policies[ip]

			sp, exist := mapPolicies[p.Name]
			if !exist {
				si.ExtractedPolicies.New = append(si.ExtractedPolicies.New, p)
				continue
			}

			sp.Roles = p.Roles
			sp.Check = p.Check
			sp.Definition = p.Definition
			si.ExtractedPolicies.Existing = append(si.ExtractedPolicies.Existing, sp)
			delete(mapPolicies, p.Name)
		}
	}

	if len(mapPolicies) > 0 {
		for _, v := range mapPolicies {
			si.ExtractedPolicies.Delete = append(si.ExtractedPolicies.Delete, v)
		}
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

func getStoragePolicies(field *reflect.StructField, si *ExtractStorageItem) (policies []objects.Policy) {
	acl := raiden.UnmarshalAclTag(string(field.Tag))

	storageType := strings.ToLower(string(supabase.RlsTypeStorage))

	var defaultCheck = fmt.Sprintf("bucket_id = '%s'", si.Storage.Name)
	var defaultDefinition = defaultCheck

	if len(acl.Read.Roles) > 0 {
		readPolicyName := supabase.GetPolicyName(objects.PolicyCommandSelect, storageType, si.Storage.Name)
		policy := objects.Policy{
			Name:       readPolicyName,
			Schema:     supabase.DefaultStorageSchema,
			Table:      supabase.DefaultObjectTable,
			Action:     "PERMISSIVE",
			Command:    objects.PolicyCommandSelect,
			Roles:      acl.Read.Roles,
			Definition: acl.Read.Using,
		}
		if policy.Definition == "" {
			policy.Definition = fmt.Sprintf("(%s)", defaultDefinition)
		} else if policy.Definition != defaultDefinition {
			policy.Definition = fmt.Sprintf("(%s AND %s)", defaultDefinition, policy.Definition)
		}
		policies = append(policies, policy)
	}

	if len(acl.Write.Roles) > 0 {
		createPolicy := objects.Policy{
			Name:    supabase.GetPolicyName(objects.PolicyCommandInsert, storageType, si.Storage.Name),
			Schema:  supabase.DefaultStorageSchema,
			Table:   supabase.DefaultObjectTable,
			Action:  "PERMISSIVE",
			Command: objects.PolicyCommandInsert,
			Roles:   acl.Write.Roles,
			Check:   acl.Write.Check,
		}
		if createPolicy.Check == nil || (createPolicy.Check != nil && *createPolicy.Check == "") {
			check := fmt.Sprintf("(%s)", defaultCheck)
			createPolicy.Check = &check
		} else if createPolicy.Check != nil && *createPolicy.Check != "" && *createPolicy.Check != defaultCheck {
			check := fmt.Sprintf("(%s AND %s)", defaultCheck, *createPolicy.Check)
			createPolicy.Check = &check
		}

		updatePolicy := objects.Policy{
			Name:       supabase.GetPolicyName(objects.PolicyCommandUpdate, storageType, si.Storage.Name),
			Schema:     supabase.DefaultStorageSchema,
			Table:      supabase.DefaultObjectTable,
			Action:     "PERMISSIVE",
			Command:    objects.PolicyCommandUpdate,
			Roles:      acl.Write.Roles,
			Definition: acl.Write.Using,
			Check:      acl.Write.Check,
		}
		if updatePolicy.Check == nil || (updatePolicy.Check != nil && *updatePolicy.Check == "") {
			check := fmt.Sprintf("(%s)", defaultCheck)
			updatePolicy.Check = &check
		} else if updatePolicy.Check != nil && *updatePolicy.Check != "" && *updatePolicy.Check != defaultCheck {
			check := fmt.Sprintf("(%s AND %s)", defaultCheck, *updatePolicy.Check)
			updatePolicy.Check = &check
		}

		if updatePolicy.Definition == "" {
			updatePolicy.Definition = fmt.Sprintf("(%s)", defaultDefinition)
		} else if updatePolicy.Definition != defaultDefinition {
			updatePolicy.Definition = fmt.Sprintf("(%s AND %s)", defaultDefinition, updatePolicy.Definition)
		}

		deletePolicy := objects.Policy{
			Name:       supabase.GetPolicyName(objects.PolicyCommandDelete, storageType, si.Storage.Name),
			Schema:     supabase.DefaultStorageSchema,
			Table:      supabase.DefaultObjectTable,
			Action:     "PERMISSIVE",
			Command:    objects.PolicyCommandDelete,
			Roles:      acl.Write.Roles,
			Definition: acl.Write.Using,
		}
		if deletePolicy.Definition == "" {
			deletePolicy.Definition = fmt.Sprintf("(%s)", defaultDefinition)
		} else if deletePolicy.Definition != defaultDefinition {
			deletePolicy.Definition = fmt.Sprintf("(%s AND %s)", defaultDefinition, deletePolicy.Definition)
		}
		policies = append(policies, createPolicy, updatePolicy, deletePolicy)
	}

	return
}
