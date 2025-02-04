package raiden

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

	TypeBase struct {
	}
)

// ----- base type default function -----
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
