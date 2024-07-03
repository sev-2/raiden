package db

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func (q *Query) Eq(column string, value any) *Query {
	q.UseWhere++

	if q.EqList == nil {
		q.EqList = &[]string{}
	}

	*q.EqList = append(
		*q.EqList,
		fmt.Sprintf("%s=eq.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) Neq(column string, value any) *Query {
	q.UseWhere++

	if q.NeqList == nil {
		q.NeqList = &[]string{}
	}

	*q.NeqList = append(
		*q.NeqList,
		fmt.Sprintf("%s=neq.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) Or(column string, value any) *Query {
	q.UseWhere++

	if q.OrList == nil {
		q.OrList = &[]string{}
	}

	*q.OrList = append(
		*q.OrList,
		fmt.Sprintf("%s=or.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) Lt(column string, value any) *Query {
	q.UseWhere++

	if q.LtList == nil {
		q.LtList = &[]string{}
	}

	*q.LtList = append(
		*q.LtList,
		fmt.Sprintf("%s=lt.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) Lte(column string, value any) *Query {
	q.UseWhere++

	if q.LteList == nil {
		q.LteList = &[]string{}
	}

	*q.LteList = append(
		*q.LteList,
		fmt.Sprintf("%s=lte.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) Gt(column string, value int) *Query {
	q.UseWhere++

	if q.GtList == nil {
		q.GtList = &[]string{}
	}

	*q.GtList = append(
		*q.GtList,
		fmt.Sprintf("%s=gt.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) Gte(column string, value any) *Query {
	q.UseWhere++

	if q.GteList == nil {
		q.GteList = &[]string{}
	}

	*q.GteList = append(
		*q.GteList,
		fmt.Sprintf("%s=gte.%s", column, getStringValue(value)),
	)

	return q
}

func (q *Query) In(column string, value any) *Query {
	q.UseWhere++

	if q.InList == nil {
		q.InList = &[]string{}
	}

	strValues := strings.Join(SliceToStringSlice(value), ",")

	*q.InList = append(
		*q.InList,
		fmt.Sprintf("%s=in.(%s)", column, strValues),
	)

	return q
}

func (q *Query) Like(column string, value string) *Query {
	q.UseWhere++

	if q.LikeList == nil {
		q.LikeList = &[]string{}
	}

	value = strings.ReplaceAll(value, "%", "*")

	*q.LikeList = append(
		*q.LikeList,
		fmt.Sprintf("%s=like.%s", column, getStringWithSpace(value)),
	)

	return q
}

func (q *Query) Ilike(column string, value string) *Query {
	q.UseWhere++

	if q.IlikeList == nil {
		q.IlikeList = &[]string{}
	}

	value = strings.ReplaceAll(value, "%", "*")

	*q.IlikeList = append(
		*q.IlikeList,
		fmt.Sprintf("%s=ilike.%s", column, getStringWithSpace(value)),
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
