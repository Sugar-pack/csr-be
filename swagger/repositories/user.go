package repositories

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"golang.org/x/crypto/bcrypt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/activearea"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/role"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/user"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
)

type UserRepository interface {
	SetUserRole(ctx context.Context, userId int, roleId int) (*ent.User, error)
	CreateUser(ctx context.Context, data *models.UserRegister) (*ent.User, error)
}

const defaultRoleSlug = "user"

type userRepository struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) UserRepository {
	return &userRepository{client: client}
}

func (r *userRepository) SetUserRole(ctx context.Context, userId int, roleId int) (foundUser *ent.User, resultError error) {
	tx, err := r.client.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}

	defer func(tx *ent.Tx) {
		err := tx.Commit()
		if err != nil {
			resultError = err
			foundUser = nil
		}
	}(tx)

	user, err := tx.User.Get(ctx, userId)
	if err != nil {
		return nil, err
	}

	role, err := tx.Role.Get(ctx, roleId)
	if err != nil {
		return nil, err
	}

	foundUser, err = r.client.User.UpdateOne(user).SetRole(role).Save(ctx)
	if err != nil {
		return nil, err
	}

	return foundUser, nil
}

func DefaultUserRole(ctx context.Context, client *ent.Client) (*ent.Role, error) {
	role, err := client.Role.Query().Where(role.Slug(defaultRoleSlug)).Only(ctx)
	if err != nil {
		return nil, err
	}

	return role, nil
}

func (r *userRepository) CreateUser(ctx context.Context, data *models.UserRegister) (createdUser *ent.User, err error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*data.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("create user error, failed to generate password hash: %s", err)
	}

	var activeAreas ent.ActiveAreas
	if len(data.ActiveAreas) > 0 {
		activeAreasIds := make([]int, len(data.ActiveAreas))
		for index, areaId := range data.ActiveAreas {
			activeAreasIds[index] = int(areaId)
		}
		activeAreas, err = r.client.ActiveArea.Query().Where(activearea.IDIn(activeAreasIds...)).All(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to find active areas")
		}
		if len(activeAreas) != len(data.ActiveAreas) {
			return nil, fmt.Errorf("invalid active areas provided")
		}
	}

	userType := user.Type(*data.Type)
	passportIssueDate := time.Time(data.PassportIssueDate)
	defaultRole, err := DefaultUserRole(ctx, r.client)
	if err != nil {
		return nil, fmt.Errorf("unable to find default role, %w", err)
	}
	createdUser, err = r.client.User.
		Create().
		SetEmail(string(data.Email)).
		SetLogin(*data.Login).
		SetName(data.Name).
		SetSurname(data.Surname).
		SetPatronymic(data.Patronymic).
		SetPassportNumber(data.PassportNumber).
		SetPassportAuthority(data.PassportAuthority).
		SetPassportIssueDate(passportIssueDate).
		SetType(userType).
		SetPhone(data.PhoneNumber).
		SetOrgName(data.OrgName).
		SetWebsite(data.Website).
		SetVk(data.Vk).
		AddActiveAreas(activeAreas...).
		SetPassword(string(hashedPassword)).
		SetRole(defaultRole).
		Save(ctx)
	return
}
