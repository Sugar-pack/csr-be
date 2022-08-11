package repositories

import (
	"context"
	"errors"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/kind"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
)

type KindRepository interface {
	CreateKind(ctx context.Context, newKind models.CreateNewKind) (*ent.Kind, error)
	AllKinds(ctx context.Context, limit, offset int, orderBy, orderColumn string) ([]*ent.Kind, error)
	AllKindsTotal(ctx context.Context) (int, error)
	KindByID(ctx context.Context, id int) (*ent.Kind, error)
	DeleteKindByID(ctx context.Context, id int) error
	UpdateKind(ctx context.Context, id int, update models.PatchKind) (*ent.Kind, error)
}

var fieldsToOrderKinds = []string{
	kind.FieldID,
	kind.FieldName,
}

type kindRepository struct {
	client *ent.Client
}

func NewKindRepository(client *ent.Client) KindRepository {
	return &kindRepository{
		client: client,
	}
}

func (r *kindRepository) CreateKind(ctx context.Context, newKind models.CreateNewKind) (*ent.Kind, error) {
	return r.client.Kind.Create().
		SetName(*newKind.Name).
		SetMaxReservationUnits(*newKind.MaxReservationUnits).
		SetMaxReservationTime(*newKind.MaxReservationTime).
		Save(ctx)
}

func (r *kindRepository) AllKindsTotal(ctx context.Context) (int, error) {
	return r.client.Kind.Query().Count(ctx)
}

func (r *kindRepository) AllKinds(ctx context.Context, limit, offset int,
	orderBy, orderColumn string) ([]*ent.Kind, error) {
	if !utils.IsOrderField(orderColumn, fieldsToOrderKinds) {
		return nil, errors.New("wrong column to order by")
	}
	orderFunc, err := utils.GetOrderFunc(orderBy, orderColumn)
	if err != nil {
		return nil, err
	}
	return r.client.Kind.Query().Order(orderFunc).Limit(limit).Offset(offset).All(ctx)
}

func (r *kindRepository) KindByID(ctx context.Context, id int) (*ent.Kind, error) {
	return r.client.Kind.Query().Where(kind.ID(id)).Only(ctx)
}

func (r *kindRepository) DeleteKindByID(ctx context.Context, id int) error {
	return r.client.Kind.DeleteOneID(id).Exec(ctx)
}

func (r *kindRepository) UpdateKind(ctx context.Context, id int, update models.PatchKind) (*ent.Kind, error) {
	kindUpdate := r.client.Kind.UpdateOneID(id)
	if update.Name != "" {
		kindUpdate = kindUpdate.SetName(update.Name)
	}
	if update.MaxReservationUnits != 0 {
		kindUpdate.SetMaxReservationUnits(update.MaxReservationUnits)
	}
	if update.MaxReservationTime != 0 {
		kindUpdate.SetMaxReservationTime(update.MaxReservationTime)
	}
	return kindUpdate.Save(ctx)
}
