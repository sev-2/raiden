package state

import (
	"reflect"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type ExtractStorageResult struct {
	Existing []objects.Bucket
	New      []objects.Bucket
	Delete   []objects.Bucket
}

func ExtractStorage(storageStates []StorageState, appStorages []raiden.Bucket) (result ExtractStorageResult, err error) {
	mapStorageState := map[string]StorageState{}
	for i := range storageStates {
		s := storageStates[i]
		mapStorageState[s.Bucket.Name] = s
	}

	for _, storage := range appStorages {
		state, isStateExist := mapStorageState[storage.Name()]
		if !isStateExist {
			s := objects.Bucket{}
			BindToSupabaseStorage(&s, storage)
			result.New = append(result.New, s)
			continue
		}

		sr := BuildStorageFromState(state, storage)
		result.Existing = append(result.Existing, sr)

		delete(mapStorageState, storage.Name())
	}

	for _, state := range mapStorageState {
		result.Delete = append(result.Delete, state.Bucket)
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

func BuildStorageFromState(ss StorageState, storage raiden.Bucket) (s objects.Bucket) {
	s = ss.Bucket
	BindToSupabaseStorage(&s, storage)
	return
}

func (er ExtractStorageResult) ToDeleteFlatMap() map[string]*objects.Bucket {
	mapData := make(map[string]*objects.Bucket)

	if len(er.Delete) > 0 {
		for i := range er.Delete {
			r := er.Delete[i]
			mapData[r.Name] = &r
		}
	}

	return mapData
}
