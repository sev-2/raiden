package raiden

import "encoding/json"

const (
	DefaultTypeSchema = "public"
)

type (
	Type interface {
		Name() string
		Schema() string
		Format() string
		Enums() []string
		Attributes() []string
		Comment() *string
	}
)

// ----- base type default function -----
type TypeBase struct {
	Value any
}

func (*TypeBase) Name() string {
	return ""
}

func (*TypeBase) Schema() string {
	return DefaultTypeSchema
}

func (*TypeBase) Format() string {
	return ""
}

func (*TypeBase) Enums() []string {
	return []string{}
}

func (*TypeBase) Attributes() []string {
	return []string{}
}

func (*TypeBase) Comment() *string {
	return nil
}

func (t *TypeBase) SetValue(v any) {
	if len(t.Enums()) == 0 {
		t.Value = v
		return
	}

	var found bool
	for _, en := range t.Enums() {
		if !found && en == v {
			found = true
		}
	}

	if !found {
		return
	}

	t.Value = v
}

func (t *TypeBase) GetValue() any {
	return t.Value
}

// Implement custom JSON marshaling
func (t TypeBase) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Value)
}

// Implement custom JSON unmarshaling
func (t *TypeBase) UnmarshalJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	t.Value = v
	return nil
}

// String returns the enum as a string
func (t *TypeBase) String() string {
	if str, ok := t.Value.(string); ok {
		return str
	}
	return ""
}
