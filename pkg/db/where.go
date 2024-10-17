package db

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func (q *Query) Eq(column string, value any) *Query {

	if q.WhereAndList == nil {
		q.WhereAndList = &[]string{}
	}

	*q.WhereAndList = append(
		*q.WhereAndList,
		fmt.Sprintf("%s=eq.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) NotEq(column string, value any) *Query {

	if q.WhereAndList == nil {
		q.WhereAndList = &[]string{}
	}

	*q.WhereAndList = append(
		*q.WhereAndList,
		fmt.Sprintf("%s=not.eq.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) OrEq(column string, value any) *Query {

	if q.WhereOrList == nil {
		q.WhereOrList = &[]string{}
	}

	*q.WhereOrList = append(
		*q.WhereOrList,
		fmt.Sprintf("%s.eq.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) Neq(column string, value any) *Query {

	if q.WhereAndList == nil {
		q.WhereAndList = &[]string{}
	}

	*q.WhereAndList = append(
		*q.WhereAndList,
		fmt.Sprintf("%s=neq.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) NotNeq(column string, value any) *Query {

	if q.WhereAndList == nil {
		q.WhereAndList = &[]string{}
	}

	*q.WhereAndList = append(
		*q.WhereAndList,
		fmt.Sprintf("%s=not.neq.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) OrNeq(column string, value any) *Query {

	if q.WhereOrList == nil {
		q.WhereOrList = &[]string{}
	}

	*q.WhereOrList = append(
		*q.WhereOrList,
		fmt.Sprintf("%s.neq.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) Lt(column string, value any) *Query {

	if q.WhereAndList == nil {
		q.WhereAndList = &[]string{}
	}

	*q.WhereAndList = append(
		*q.WhereAndList,
		fmt.Sprintf("%s=lt.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) NotLt(column string, value any) *Query {

	if q.WhereAndList == nil {
		q.WhereAndList = &[]string{}
	}

	*q.WhereAndList = append(
		*q.WhereAndList,
		fmt.Sprintf("%s=not.lt.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) OrLt(column string, value any) *Query {

	if q.WhereOrList == nil {
		q.WhereOrList = &[]string{}
	}

	*q.WhereOrList = append(
		*q.WhereOrList,
		fmt.Sprintf("%s.lt.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) Lte(column string, value any) *Query {

	if q.WhereAndList == nil {
		q.WhereAndList = &[]string{}
	}

	*q.WhereAndList = append(
		*q.WhereAndList,
		fmt.Sprintf("%s=lte.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) NotLte(column string, value any) *Query {

	if q.WhereAndList == nil {
		q.WhereAndList = &[]string{}
	}

	*q.WhereAndList = append(
		*q.WhereAndList,
		fmt.Sprintf("%s=not.lte.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) OrLte(column string, value any) *Query {

	if q.WhereOrList == nil {
		q.WhereOrList = &[]string{}
	}

	*q.WhereOrList = append(
		*q.WhereOrList,
		fmt.Sprintf("%s.lte.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) Gt(column string, value int) *Query {

	if q.WhereAndList == nil {
		q.WhereAndList = &[]string{}
	}

	*q.WhereAndList = append(
		*q.WhereAndList,
		fmt.Sprintf("%s=gt.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) NotGt(column string, value int) *Query {

	if q.WhereAndList == nil {
		q.WhereAndList = &[]string{}
	}

	*q.WhereAndList = append(
		*q.WhereAndList,
		fmt.Sprintf("%s=not.gt.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) OrGt(column string, value int) *Query {

	if q.WhereOrList == nil {
		q.WhereOrList = &[]string{}
	}

	*q.WhereOrList = append(
		*q.WhereOrList,
		fmt.Sprintf("%s.gt.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) Gte(column string, value any) *Query {

	if q.WhereAndList == nil {
		q.WhereAndList = &[]string{}
	}

	*q.WhereAndList = append(
		*q.WhereAndList,
		fmt.Sprintf("%s=gte.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) NotGte(column string, value any) *Query {

	if q.WhereAndList == nil {
		q.WhereAndList = &[]string{}
	}

	*q.WhereAndList = append(
		*q.WhereAndList,
		fmt.Sprintf("%s=not.gte.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) OrGte(column string, value any) *Query {

	if q.WhereOrList == nil {
		q.WhereOrList = &[]string{}
	}

	*q.WhereOrList = append(
		*q.WhereOrList,
		fmt.Sprintf("%s.gte.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) In(column string, value any) *Query {

	if q.WhereAndList == nil {
		q.WhereAndList = &[]string{}
	}

	strValues := strings.Join(SliceToStringSlice(value), ",")

	*q.WhereAndList = append(
		*q.WhereAndList,
		fmt.Sprintf("%s=in.(%s)", column, strValues),
	)

	return q
}

func (q *Query) NotIn(column string, value any) *Query {

	if q.WhereAndList == nil {
		q.WhereAndList = &[]string{}
	}

	strValues := strings.Join(SliceToStringSlice(value), ",")

	*q.WhereAndList = append(
		*q.WhereAndList,
		fmt.Sprintf("%s=not.in.(%s)", column, strValues),
	)

	return q
}

func (q *Query) OrIn(column string, value any) *Query {

	if q.WhereOrList == nil {
		q.WhereOrList = &[]string{}
	}

	strValues := strings.Join(SliceToStringSlice(value), ",")

	*q.WhereOrList = append(
		*q.WhereOrList,
		fmt.Sprintf("%s.in.(%s)", column, strValues),
	)

	return q
}

func (q *Query) Like(column string, value string) *Query {

	if q.WhereAndList == nil {
		q.WhereAndList = &[]string{}
	}

	value = strings.ReplaceAll(value, "%", "*")

	*q.WhereAndList = append(
		*q.WhereAndList,
		fmt.Sprintf("%s=like.%s", column, getStringWithSpace(value)),
	)

	return q
}

func (q *Query) NotLike(column string, value string) *Query {

	if q.WhereAndList == nil {
		q.WhereAndList = &[]string{}
	}

	value = strings.ReplaceAll(value, "%", "*")

	*q.WhereAndList = append(
		*q.WhereAndList,
		fmt.Sprintf("%s=not.like.%s", column, getStringWithSpace(value)),
	)

	return q
}

func (q *Query) OrLike(column string, value string) *Query {

	if q.WhereOrList == nil {
		q.WhereOrList = &[]string{}
	}

	value = strings.ReplaceAll(value, "%", "*")

	*q.WhereOrList = append(
		*q.WhereOrList,
		fmt.Sprintf("%s.like.%s", column, getStringWithSpace(value)),
	)

	return q
}

func (q *Query) Ilike(column string, value string) *Query {

	if q.WhereAndList == nil {
		q.WhereAndList = &[]string{}
	}

	value = strings.ReplaceAll(value, "%", "*")

	*q.WhereAndList = append(
		*q.WhereAndList,
		fmt.Sprintf("%s=ilike.%s", column, getStringWithSpace(value)),
	)

	return q
}

func (q *Query) NotIlike(column string, value string) *Query {

	if q.WhereAndList == nil {
		q.WhereAndList = &[]string{}
	}

	value = strings.ReplaceAll(value, "%", "*")

	*q.WhereAndList = append(
		*q.WhereAndList,
		fmt.Sprintf("%s=not.ilike.%s", column, getStringWithSpace(value)),
	)

	return q
}

func (q *Query) OrIlike(column string, value string) *Query {

	if q.WhereOrList == nil {
		q.WhereOrList = &[]string{}
	}

	value = strings.ReplaceAll(value, "%", "*")

	*q.WhereOrList = append(
		*q.WhereOrList,
		fmt.Sprintf("%s.ilike.%s", column, getStringWithSpace(value)),
	)

	return q
}

func (q *Query) Is(column string, value any) *Query {

	if !isValueWhitelist(value) {
		panic("isValueWhitelist: only \"true\", \"false\", \"null\", or \"unknown\" are allowed")
	}

	if q.IsList == nil {
		q.IsList = &[]string{}
	}

	*q.IsList = append(
		*q.IsList,
		fmt.Sprintf("%s=is.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) NotIs(column string, value any) *Query {

	if !isValueWhitelist(value) {
		panic("isValueWhitelist: only \"true\", \"false\", \"null\", or \"unknown\" are allowed")
	}

	if q.IsList == nil {
		q.IsList = &[]string{}
	}

	*q.IsList = append(
		*q.IsList,
		fmt.Sprintf("%s=not.is.%s", column, getStringValue(value)),
	)

	return q
}

func getStringValue(value any) string {
	return fmt.Sprintf("%v", value)
}

func getStringWithSpace(value string) string {
	return strings.ReplaceAll(value, " ", "%20")
}

func SliceToStringSlice(slice interface{}) []string {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		panic("SliceToStringSlice: not a slice")
	}

	stringSlice := make([]string, v.Len())
	for i := 0; i < v.Len(); i++ {
		switch v.Index(i).Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			stringSlice[i] = strconv.FormatInt(v.Index(i).Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			stringSlice[i] = strconv.FormatUint(v.Index(i).Uint(), 10)
		case reflect.Float32, reflect.Float64:
			stringSlice[i] = strconv.FormatFloat(v.Index(i).Float(), 'f', -1, 64)
		case reflect.Bool:
			stringSlice[i] = strconv.FormatBool(v.Index(i).Bool())
		case reflect.String:
			stringSlice[i] = v.Index(i).String()
		default:
			panic("SliceToStringSlice: unsupported slice element type")
		}
	}
	return stringSlice
}

func isValueWhitelist(value any) bool {
	switch getStringValue(value) {
	case "true":
		return true
	case "false":
		return true
	case "null":
		return true
	case "unknown":
		return true
	default:
		return false
	}
}
