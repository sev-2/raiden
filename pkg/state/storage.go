package state

import (
	"reflect"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type ExtractStorageResult struct {
	Existing []objects.Storage
	New      []objects.Storage
	Delete   []objects.Storage
}

func ExtractStorage(storageStates []StorageState, appStorages []raiden.Storage) (result ExtractStorageResult, err error) {
	mapStorageState := map[string]StorageState{}
	for i := range storageStates {
		s := storageStates[i]
		mapStorageState[s.Storage.Name] = s
	}

	for _, storage := range appStorages {
		state, isStateExist := mapStorageState[storage.Name()]
		if !isStateExist {
			s := objects.Storage{}
			BindToSupabaseStorage(&s, storage)
			result.New = append(result.New, s)
			continue
		}

		sr := BuildStorageFromState(state, storage)
		result.Existing = append(result.Existing, sr)

		delete(mapStorageState, storage.Name())
	}

	for _, state := range mapStorageState {
		result.Delete = append(result.Delete, state.Storage)
	}

	return
}

func BindToSupabaseStorage(s *objects.Storage, storage raiden.Storage) {
	name := storage.Name()
	if name == "" {
		rv := reflect.TypeOf(storage)
		name = utils.ToSnakeCase(rv.Name())
	}

	s.Name = name
	s.Public = storage.Public()
	s.AllowedMimeTypes = storage.AllowedMimeTypes()
	s.AvifAutoDetection = storage.AvifAutoDetection()
	s.FileSizeLimit = storage.FileSizeLimit()
}

func BuildStorageFromState(ss StorageState, storage raiden.Storage) (s objects.Storage) {
	s = ss.Storage
	BindToSupabaseStorage(&s, storage)
	return
}
