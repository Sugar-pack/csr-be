package utils

import (
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
)

const (
	AscOrder  = "asc"
	DescOrder = "desc"
)

func GetOrderFunc(orderBy, orderColumn string) (func(*sql.Selector), error) {
	var orderFunc func(*sql.Selector)
	var err error
	switch orderBy {
	case AscOrder:
		orderFunc = ent.Asc(orderColumn)
	case DescOrder:
		orderFunc = ent.Desc(orderColumn)
	default:
		err = errors.New(fmt.Sprintf("wrong value for orderBy: %s", orderBy))
	}
	return orderFunc, err
}
