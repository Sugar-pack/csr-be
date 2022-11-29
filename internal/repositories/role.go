package repositories

import (
	"context"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/role"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

type roleRepository struct {
}

func (r *roleRepository) GetRoles(ctx context.Context) ([]*ent.Role, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return tx.Role.Query().Order(ent.Asc(role.FieldID)).All(ctx)
}

func NewRoleRepository() domain.RoleRepository {
	return &roleRepository{}
}
