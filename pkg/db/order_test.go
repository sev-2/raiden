package db

import (
	"testing"

	"github.com/sev-2/raiden"
)

func TestOrderAsc(t *testing.T) {
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).OrderAsc("age").OrderAsc("name")

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
	ctx := raiden.Ctx{}

	q := NewQuery(&ctx).OrderDesc("age").OrderDesc("name")

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
