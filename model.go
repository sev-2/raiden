package raiden

import (
	"regexp"
	"strings"
)

type (
	// definition of column tag, example :
	// column:"name:id;type:bigint;primaryKey;autoIncrement;nullable:false;unique;default:now()"
	ColumnTag struct {
		Name          string
		Type          string
		PrimaryKey    bool
		AutoIncrement bool
		Nullable      bool
		Default       *string
		Unique        bool
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

func MarshalColumnTag(tag string) ColumnTag {
	columnTag := ColumnTag{
		Nullable: true,
	}

	// Regular expression to match key-value pairs
	re := regexp.MustCompile(`(\w+):([^;]+);?`)

	// Find all matches in the tag string
	matches := re.FindAllStringSubmatch(tag, -1)

	// Iterate over matches and populate ColumnTag fields
	for _, match := range matches {
		key := match[1]
		value := strings.TrimSpace(match[2])

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
				columnTag.Default = &value
			}
		case "unique":
			columnTag.Unique = true
		}
	}

	return columnTag
}

func MarshallJoinTag(tag string) JoinTag {
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
