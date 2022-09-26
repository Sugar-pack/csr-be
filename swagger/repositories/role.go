package repositories

import (
	"context"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/role"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
)

type RoleRepository interface {
	GetRoles(ctx context.Context) ([]*ent.Role, error)
}

type roleRepository struct {
}

func (r *roleRepository) GetRoles(ctx context.Context) ([]*ent.Role, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return tx.Role.Query().Order(ent.Asc(role.FieldID)).All(ctx)
}

func NewRoleRepository() RoleRepository {
	return &roleRepository{}
}
