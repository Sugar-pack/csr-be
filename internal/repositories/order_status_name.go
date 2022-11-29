package repositories

import (
	"context"
	"fmt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
)

type orderStatusNameRepository struct {
}

func NewOrderStatusNameRepository() *orderStatusNameRepository {
	return &orderStatusNameRepository{}
}

func (r *orderStatusNameRepository) ListOfOrderStatusNames(ctx context.Context) ([]*ent.OrderStatusName, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	pointersStatuses, err := tx.OrderStatusName.Query().Order(ent.Asc("id")).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("status history error, failed to get status names: %s", err)
	}
	return pointersStatuses, nil
}
