package objects

type TypeAttribute struct {
	Name   string `json:"name"`
	TypeID int    `json:"type_id"`
}

type Type struct {
	ID         int             `json:"id"`
	Name       string          `json:"name"`
	Schema     string          `json:"schema"`
	Format     string          `json:"format"`
	Enums      []string        `json:"enums"`
	Attributes []TypeAttribute `json:"attributes"`
	Comment    *string         `json:"comment"` // Use pointer to handle null values
}

type UpdateDataType string

const (
	UpdateTypeName       UpdateDataType = "name"
	UpdateTypeSchema     UpdateDataType = "schema"
	UpdateTypeFormat     UpdateDataType = "format"
	UpdateTypeEnums      UpdateDataType = "enums"
	UpdateTypeAttributes UpdateDataType = "attributes"
	UpdateTypeComment    UpdateDataType = "comment"
)

type UpdateTypeParam struct {
	OldData     Type
	ChangeItems []UpdateDataType
}
