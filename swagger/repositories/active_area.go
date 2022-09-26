package repositories

import (
	"context"
	"errors"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/activearea"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
)

type ActiveAreaRepository interface {
	AllActiveAreas(ctx context.Context, limit, offset int, orderBy, orderColumn string) ([]*ent.ActiveArea, error)
	TotalActiveAreas(ctx context.Context) (int, error)
}

var fieldsToOrderAreas = []string{
	activearea.FieldID,
	activearea.FieldName,
}

type activeAreaRepository struct {
}

func NewActiveAreaRepository() ActiveAreaRepository {
	return &activeAreaRepository{}
}

func (r *activeAreaRepository) AllActiveAreas(ctx context.Context, limit, offset int,
	orderBy, orderColumn string) ([]*ent.ActiveArea, error) {
	if !utils.IsOrderField(orderColumn, fieldsToOrderAreas) {
		return nil, errors.New("wrong column to order by")
	}
	orderFunc, err := utils.GetOrderFunc(orderBy, orderColumn)
	if err != nil {
		return nil, err
	}
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return tx.ActiveArea.Query().Order(orderFunc).Limit(limit).Offset(offset).All(ctx)
}

func (r *activeAreaRepository) TotalActiveAreas(ctx context.Context) (int, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return 0, err
	}
	return tx.ActiveArea.Query().Count(ctx)
}
