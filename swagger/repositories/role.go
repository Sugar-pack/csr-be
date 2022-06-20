package repositories

import (
	"context"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/role"
)

type RoleRepository interface {
	GetRoles(ctx context.Context) ([]*ent.Role, error)
}

type roleRepository struct {
	client *ent.Client
}

func (r *roleRepository) GetRoles(ctx context.Context) ([]*ent.Role, error) {
	return r.client.Role.Query().Order(ent.Asc(role.FieldID)).All(ctx)
}

func NewRoleRepository(client *ent.Client) RoleRepository {
	return &roleRepository{
		client: client,
	}
}
