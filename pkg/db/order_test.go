package db

import (
	"testing"
)

func TestOrderAsc(t *testing.T) {
	q := NewQuery(&mockRaidenContext).OrderAsc("age").OrderAsc("name")

	orderList := *q.OrderList

	if orderList == nil {
		t.Error("Expected order not to be nil")
	}

	if len(orderList) != 2 {
		t.Error("Expected have 2 order clause")
	}

	if orderList[0] != "age.asc" {
		t.Errorf("Expected the first order clause index is \"%s\"", "age.asc")
	}

	if orderList[1] != "name.asc" {
		t.Errorf("Expected the second order clause index is \"%s\"", "name.asc")
	}
}

func TestOrderDesc(t *testing.T) {
	q := NewQuery(&mockRaidenContext).OrderDesc("age").OrderDesc("name")

	orderList := *q.OrderList

	if orderList == nil {
		t.Error("Expected order not to be nil")
	}

	if len(orderList) != 2 {
		t.Error("Expected have 2 order clause")
	}

	if orderList[0] != "age.desc" {
		t.Errorf("Expected the first order clause index is \"%s\"", "age.asc")
	}

	if orderList[1] != "name.desc" {
		t.Errorf("Expected the second order clause index is \"%s\"", "name.asc")
	}
}
