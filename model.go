package raiden

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/sev-2/raiden/pkg/utils"
)

type (
	ModelBase struct {
	}

	// definition of column tag, example :
	// column:"name:id;type:bigint;primaryKey;autoIncrement;nullable:false;unique;default:now()"
	ColumnTag struct {
		Name          string
		Type          string
		PrimaryKey    bool
		AutoIncrement bool
		Nullable      bool
		Default       any
		Unique        bool
		Index         bool
	}

	// definition of join tag, example:
	// - join:"joinType:hasOne;primaryKey:id;foreignKey:candidate_id"
	// - join:"joinType:hasMany;primaryKey:id;foreignKey:scouter_id"
	// - join:"joinType:manyToMany;through:submission;sourcePrimaryKey:id;sourceForeignKey:candidate_id;targetPrimaryKey:id;targetForeign:candidate_id"
	JoinTag struct {
		JoinType   RelationType
		PrimaryKey string
		ForeignKey string

		Through          string
		SourcePrimaryKey string
		SourceForeignKey string

		TargetPrimaryKey string
		TargetForeignKey string
	}

	RelationType string
)

var (
	RelationTypeHasOne     RelationType = "hasOne"
	RelationTypeHasMany    RelationType = "hasMany"
	RelationTypeManyToMany RelationType = "manyToMany"
)

func UnmarshalColumnTag(tag string) ColumnTag {
	columnTag := ColumnTag{
		Nullable:      true,
		PrimaryKey:    false,
		AutoIncrement: false,
		Unique:        false,
	}

	tagSplit := strings.Split(tag, ";")
	tagMap := make(map[string]string)
	for _, c := range tagSplit {
		cSplit := strings.Split(c, ":")
		if len(cSplit) == 2 {
			tagMap[cSplit[0]] = cSplit[1]
		} else if len(cSplit) == 1 {
			tagMap[cSplit[0]] = ""
		}
	}

	defaultValue := "default"

	// Iterate over matches and populate ColumnTag fields
	for key, value := range tagMap {
		switch key {
		case "name":
			columnTag.Name = value
		case "type":
			columnTag.Type = value
		case "primaryKey":
			columnTag.PrimaryKey = true
		case "autoIncrement":
			columnTag.AutoIncrement = true
		case "nullable":
			if value != "" {
				columnTag.Nullable = value == "true"
			}
		case "default":
			if value != "" {
				defaultValue = value
			}
		case "unique":
			columnTag.Unique = true
		}
	}

	if defaultValue != "default" {
		columnTag.Default = &defaultValue
	}

	return columnTag
}

func UnmarshalJoinTag(tag string) JoinTag {
	joinTag := JoinTag{}

	// Regular expression to match key-value pairs
	re := regexp.MustCompile(`(\w+):([^;]+);?`)

	// Find all matches in the tag string
	matches := re.FindAllStringSubmatch(tag, -1)

	// Iterate over matches and populate JoinTag fields
	for _, match := range matches {
		key := match[1]
		value := match[2]

		switch key {
		case "joinType":
			joinTag.JoinType = RelationType(value)
		case "primaryKey":
			joinTag.PrimaryKey = value
		case "foreignKey":
			joinTag.ForeignKey = value
		case "through":
			joinTag.Through = value
		case "sourcePrimaryKey":
			joinTag.SourcePrimaryKey = value
		case "sourceForeignKey":
			joinTag.SourceForeignKey = value
		case "targetPrimaryKey":
			joinTag.TargetPrimaryKey = value
		case "targetForeign":
			joinTag.TargetForeignKey = value
		}
	}

	return joinTag
}

func GetTableName(model any) (tableName string) {
	rt := reflect.TypeOf(model)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	tableName = strings.ToLower(utils.ToSnakeCase(rt.Name()))
	field, found := rt.FieldByName("Metadata")
	if found {
		foundTableName := field.Tag.Get("tableName")
		if foundTableName != "" {
			tableName = foundTableName
		}

	}
	return
}
