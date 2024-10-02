package db

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sev-2/raiden"
)

func (q *Query) Preload(table string, args ...string) *Query {

	relatedFieldPrefix := ""
	relationMap := make(map[string]map[string]string)

	field := ""
	operator := ""
	value := ""

	// Override with supplied arguments if available
	if len(args) > 0 && args[0] != "" {
		field = args[0]
	}

	if len(args) > 1 && args[1] != "" {
		operator = args[1]
	}
	if len(args) > 2 && args[2] != "" {
		value = args[2]
	}

	relations := strings.Split(table, ".")

	fmt.Printf("Preloading table: %s, field: %s, operator: %s, value: %s\n", table, field, operator, value)

	if len(relations) > 3 {
		raiden.Fatal("unsupported nested relations more than 3 levels")
	}

	for i, relation := range relations {
		var currentModelStruct reflect.Type
		var relatedModel interface{}
		var err error
		if i == 0 {
			currentModelStruct = reflect.TypeOf(q.model)
			relatedModel, err = instantiateFieldByPath(q.model, relation)
		} else {
			currentModelStruct = reflect.TypeOf(relatedModel)
			relatedModel, err = instantiateFieldByPath(relatedModel, relation)
		}

		if err != nil {
			raiden.Fatal("could not find related model.")
		}

		fmt.Printf("Related model: %v\n", relatedModel)
		relatedModelStruct := reflect.TypeOf(relatedModel)
		if relatedModelStruct.Kind() == reflect.Ptr {
			relatedModelStruct = relatedModelStruct.Elem()
		}

		var relatedAlias string
		var relatedTableName string
		var relatedForeignKey string
		for i := 0; i < relatedModelStruct.NumField(); i++ {
			field := relatedModelStruct.Field(i)
			if field.Name == "Metadata" {
				relatedTableName = field.Tag.Get("tableName")
			}
		}

		if currentModelStruct.Kind() == reflect.Ptr {
			currentModelStruct = currentModelStruct.Elem()
		}

		for i := 0; i < currentModelStruct.NumField(); i++ {
			field := currentModelStruct.Field(i)
			if field.Name == relation {
				jsonField := field.Tag.Get("json")
				join := field.Tag.Get("join")

				relatedAlias = strings.Split(jsonField, ",")[0]
				relatedForeignKey, err = getTagValue(join, "foreignKey")

				if err != nil {
					raiden.Fatal("could not find foreign key in join tag.")
				}
			}
		}

		relationData := make(map[string]string)
		relationData["alias"] = relatedAlias
		relationData["table"] = relatedTableName
		relationData["fk"] = relatedForeignKey
		relationMap[relation] = relationData
	}

	var selects []string

	// After we have the relation map, we can construct the select query
	// If the table is `Users.Team.Organization`,
	// the select query will be `users(teams(organizations(*)))`
	for _, r := range reverseSortString(relations) {
		d := relationMap[r]
		alias := d["alias"]
		table := d["table"]
		fk := d["fk"]

		var related string
		if alias == table {
			related = fmt.Sprintf("%s!%s", table, fk)
		} else {
			related = fmt.Sprintf("%s:%s!%s", alias, table, fk)
		}

		if len(selects) > 0 {
			lastQuery := selects[len(selects)-1]
			selects[len(selects)-1] = fmt.Sprintf("%s(%s,%s)", related, "*", lastQuery)
		} else {
			selects = append(selects, fmt.Sprintf("%s(%s)", related, "*"))
		}

		if relatedFieldPrefix == "" {
			relatedFieldPrefix = table
		} else {
			relatedFieldPrefix = fmt.Sprintf("%s.%s", relatedFieldPrefix, table)
		}
	}

	fmt.Println("Relations: ", relationMap)
	fmt.Println("Selects: ", selects)
	fmt.Println("Prefix: ", relatedFieldPrefix)
	q.Relations = append(q.Relations, selects...)

	if field != "" && operator != "" && value != "" {
		if q.WhereAndList == nil {
			q.WhereAndList = &[]string{}
		}

		*q.WhereAndList = append(
			*q.WhereAndList,
			fmt.Sprintf("%s=%s.%s", fmt.Sprintf("%s.%s", relatedFieldPrefix, field), operator, getStringValue(value)),
		)
	}

	return q
}

func instantiateFieldByPath(model interface{}, fieldPath string) (interface{}, error) {
	fields := strings.Split(fieldPath, ".")
	val := reflect.ValueOf(model)

	// If it's a pointer, dereference it, but keep track of the original value to modify it
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct, got %s", val.Kind())
	}

	// Traverse the struct fields based on the field path
	for _, fieldName := range fields {
		fieldVal := val.FieldByName(fieldName)

		if !fieldVal.IsValid() {
			return nil, fmt.Errorf("field %s not found", fieldName)
		}

		if fieldVal.Kind() == reflect.Ptr {
			if fieldVal.IsNil() {
				fieldType := fieldVal.Type()

				fieldVal = reflect.New(fieldType.Elem())
			}
			fieldVal = fieldVal.Elem()
		}

		if fieldVal.Kind() != reflect.Struct {
			return nil, fmt.Errorf("field %s is not a struct", fieldName)
		}

		val = fieldVal
	}

	newInstance := reflect.New(val.Type()).Elem().Interface()
	return newInstance, nil
}

func getTagValue(tag, key string) (string, error) {
	// Split the tag by semicolon to get individual key-value pairs
	pairs := strings.Split(tag, ";")

	// Iterate through the pairs to find the key
	for _, pair := range pairs {
		// Split the pair by colon to get key and value
		kv := strings.Split(pair, ":")
		if len(kv) == 2 && kv[0] == key {
			return kv[1], nil // Return the value if key matches
		}
	}

	return "", fmt.Errorf("key %s not found in tag", key)
}
