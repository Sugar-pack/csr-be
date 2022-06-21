package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/activearea"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/role"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/user"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
)

type UserRepository interface {
	SetUserRole(ctx context.Context, userId int, roleId int) error
	UserByLogin(ctx context.Context, login string) (*ent.User, error)
	ChangePasswordByLogin(ctx context.Context, login string, password string) (Transaction, error)
	CreateUser(ctx context.Context, data *models.UserRegister) (*ent.User, error)
	GetUserByLogin(ctx context.Context, login string) (*ent.User, error)
	GetUserByID(ctx context.Context, id int) (*ent.User, error)
	UpdateUserByID(ctx context.Context, id int, patch *models.PatchUserRequest) error
	UserList(ctx context.Context) ([]*ent.User, error)
	ConfirmRegistration(ctx context.Context, login string) error
}

const (
	defaultRoleSlug   = "user"
	defaultStrFmtDate = "0001-01-01"
)

type userRepository struct {
	client *ent.Client
}

func (r *userRepository) UserList(ctx context.Context) ([]*ent.User, error) {
	return r.client.User.Query().WithRole().All(ctx)
}

func (r *userRepository) UpdateUserByID(ctx context.Context, id int, patch *models.PatchUserRequest) error {
	if patch == nil {
		return errors.New("patch is nil")
	}
	userUpdate := r.client.User.Update().Where(user.ID(id))
	if patch.Name != "" {
		userUpdate.SetName(patch.Name)
	}
	if patch.Surname != "" {
		userUpdate.SetName(patch.Name)
	}
	if patch.PassportNumber != "" {
		//TODO: if user changes his passport data or name/surname,
		// we should recall him to verify it
		userUpdate.SetPassportNumber(patch.PassportNumber)
	}
	if patch.PassportAuthority != "" {
		userUpdate.SetPassportAuthority(patch.PassportAuthority)
	}
	if patch.PassportIssueDate.String() != defaultStrFmtDate { // todo think how to do it right
		passportIssueDate := time.Time(patch.PassportIssueDate)
		userUpdate.SetPassportIssueDate(passportIssueDate)
	}
	if patch.OrgName != "" {
		userUpdate.SetOrgName(patch.OrgName)
	}
	if patch.Vk != "" {
		userUpdate.SetVk(patch.Vk)
	}
	if patch.Website != "" {
		userUpdate.SetWebsite(patch.Website)
	}

	_, err := userUpdate.Save(ctx)
	return err
}

func (r *userRepository) GetUserByLogin(ctx context.Context, login string) (*ent.User, error) {
	return r.client.User.Query().Where(user.Login(login)).WithGroups().WithRole().Only(ctx)
}

func (r *userRepository) UserByLogin(ctx context.Context, login string) (*ent.User, error) {
	return r.client.User.Query().Where(user.LoginEQ(login)).Only(ctx)
}

func (r *userRepository) ChangePasswordByLogin(ctx context.Context, login string, password string) (Transaction, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	_, err = tx.User.Update().Where(user.LoginEQ(login)).SetPassword(password).Save(ctx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func NewUserRepository(client *ent.Client) UserRepository {
	return &userRepository{client: client}
}

func (r *userRepository) GetUserByID(ctx context.Context, id int) (*ent.User, error) {
	return r.client.User.Query().Where(user.ID(id)).WithGroups().WithRole().Only(ctx)
}

func (r *userRepository) SetUserRole(ctx context.Context, userId int, roleId int) error {
	_, err := r.client.User.UpdateOneID(userId).SetRoleID(roleId).Save(ctx)
	if err != nil {
		return err
	}
	return nil
}

func DefaultUserRole(ctx context.Context, client *ent.Client) (*ent.Role, error) {
	defaultRole, err := client.Role.Query().Where(role.Slug(defaultRoleSlug)).Only(ctx)
	if err != nil {
		return nil, err
	}

	return defaultRole, nil
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

func (r *userRepository) ConfirmRegistration(ctx context.Context, login string) error {

	_, err := r.client.User.Update().Where(user.LoginEQ(login)).SetIsConfirmed(true).Save(ctx)
	if err != nil {
		return err
	}
	return nil
}
