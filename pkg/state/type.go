package state

import (
	"reflect"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type ExtractTypeResult struct {
	Existing []objects.Type
	New      []objects.Type
	Delete   []objects.Type
}

func ExtractType(typeStates []TypeState, appTypes []raiden.Type) (result ExtractTypeResult, err error) {
	mapTypeState := map[string]TypeState{}
	for i := range typeStates {
		r := typeStates[i]
		mapTypeState[r.Type.Name] = r
	}

	for _, dataType := range appTypes {
		state, isStateExist := mapTypeState[dataType.Name()]
		if !isStateExist {
			t := objects.Type{}
			BindToSupabaseType(&t, dataType)
			result.New = append(result.New, t)
			continue
		}

		sr := BuildTypeFromState(state, dataType)
		result.Existing = append(result.Existing, sr)
		delete(mapTypeState, dataType.Name())
	}

	for _, state := range mapTypeState {
		result.Delete = append(result.Delete, state.Type)
	}

	return
}

func BindToSupabaseType(r *objects.Type, role raiden.Type) {
	name := role.Name()
	if name == "" {
		rv := reflect.TypeOf(role)
		name = utils.ToSnakeCase(rv.Name())
	}

	r.Name = name
	r.Attributes = role.Attributes()
	r.Comment = role.Comment()
	r.Enums = role.Enums()
	r.Format = role.Format()
	r.Schema = role.Schema()
}

func BuildTypeFromState(ts TypeState, t raiden.Type) (r objects.Type) {
	r = ts.Type
	BindToSupabaseType(&r, t)
	return
}

func (er ExtractTypeResult) ToDeleteFlatMap() map[string]*objects.Type {
	mapData := make(map[string]*objects.Type)

	if len(er.Delete) > 0 {
		for i := range er.Delete {
			r := er.Delete[i]
			mapData[r.Name] = &r
		}
	}

	return mapData
}

func (er ExtractTypeResult) ToMap() map[string]objects.Type {
	mapData := make(map[string]objects.Type)

	if len(er.New) > 0 {
		for i := range er.New {
			r := er.New[i]
			mapData[r.Name] = r
		}
	}

	if len(er.Existing) > 0 {
		for i := range er.Existing {
			r := er.Existing[i]
			mapData[r.Name] = r
		}
	}

	return mapData
}
