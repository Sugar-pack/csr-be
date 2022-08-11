package repositories

import (
	"context"
	"errors"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/activearea"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
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
	client *ent.Client
}

func NewActiveAreaRepository(client *ent.Client) ActiveAreaRepository {
	return &activeAreaRepository{client: client}
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
	return r.client.ActiveArea.Query().Order(orderFunc).Limit(limit).Offset(offset).All(ctx)
}

func (r *activeAreaRepository) TotalActiveAreas(ctx context.Context) (int, error) {
	return r.client.ActiveArea.Query().Count(ctx)
}
