package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/activearea"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/role"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/user"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
)

const (
	defaultRoleSlug   = "user"
	defaultStrFmtDate = "0001-01-01"
)

var fieldsToOrderUsers = []string{
	user.FieldID,
	user.FieldName,
	user.FieldLogin,
	user.FieldEmail,
}

type userRepository struct {
}

func (r *userRepository) UsersListTotal(ctx context.Context) (int, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return 0, err
	}
	return tx.User.Query().Where(user.IsDeleted(false)).Count(ctx)
}

func (r *userRepository) UserList(ctx context.Context, limit, offset int,
	orderBy, orderColumn string) ([]*ent.User, error) {
	if !utils.IsValueInList(orderColumn, fieldsToOrderUsers) {
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
	return tx.User.Query().Where(user.IsDeleted(false)).WithRole().Order(orderFunc).Limit(limit).Offset(offset).All(ctx)
}

func (r *userRepository) UpdateUserByID(ctx context.Context, id int, patch *models.PatchUserRequest) error {
	if patch == nil {
		return errors.New("patch is nil")
	}
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	userUpdate := tx.User.Update().Where(user.ID(id))
	if patch.Name != "" {
		userUpdate.SetName(patch.Name)
	}
	if patch.Surname != "" {
		userUpdate.SetSurname(patch.Surname)
	}
	if patch.PassportSeries != "" {
		userUpdate.SetPassportSeries(patch.PassportSeries)
	}
	if patch.Phone != "" {
		userUpdate.SetPhone(patch.Phone)
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

	_, err = userUpdate.Save(ctx)
	return err
}

func (r *userRepository) GetUserByLogin(ctx context.Context, login string) (*ent.User, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return tx.User.Query().Where(user.Login(login)).WithGroups().WithRole().WithRegistrationConfirm().Only(ctx)
}

func (r *userRepository) UserByLogin(ctx context.Context, login string) (*ent.User, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return tx.User.Query().Where(user.LoginEQ(login)).Only(ctx)
}

func (r *userRepository) ChangePasswordByLogin(ctx context.Context, login string, password string) error {
	hash, err := utils.PasswordHash(password)
	if err != nil {
		return fmt.Errorf("error while hashing password: %w", err)
	}

	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	_, err = tx.User.Update().Where(user.LoginEQ(login)).SetPassword(hash).Save(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) ChangeEmailByLogin(ctx context.Context, login string, email string) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}

	_, err = tx.User.Update().Where(user.LoginEQ(login)).SetEmail(email).Save(ctx)
	if err != nil {
		return err
	}
	return nil
}

func NewUserRepository() domain.UserRepository {
	return &userRepository{}
}

func (r *userRepository) GetUserByID(ctx context.Context, id int) (*ent.User, error) {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return tx.User.Query().Where(user.ID(id)).WithGroups().WithRole().Only(ctx)
}

func (r *userRepository) SetUserRole(ctx context.Context, userId int, roleId int) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	_, err = tx.User.UpdateOneID(userId).SetRoleID(roleId).Save(ctx)
	if err != nil {
		return err
	}
	return nil
}

func DefaultUserRole(ctx context.Context, tx *ent.Tx) (*ent.Role, error) {
	defaultRole, err := tx.Role.Query().Where(role.Slug(defaultRoleSlug)).Only(ctx)
	if err != nil {
		return nil, err
	}

	return defaultRole, nil
}

func (r *userRepository) CreateUser(ctx context.Context, data *models.UserRegister) (createdUser *ent.User, err error) {
	hashedPassword, err := utils.PasswordHash(*data.Password)
	if err != nil {
		return nil, fmt.Errorf("create user error, failed to generate password hash: %s", err)
	}

	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var activeAreas ent.ActiveAreas
	if len(data.ActiveAreas) > 0 {
		activeAreasIds := make([]int, len(data.ActiveAreas))
		for index, areaId := range data.ActiveAreas {
			activeAreasIds[index] = int(areaId)
		}
		activeAreas, err = tx.ActiveArea.Query().Where(activearea.IDIn(activeAreasIds...)).All(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to find active areas")
		}
		if len(activeAreas) != len(data.ActiveAreas) {
			return nil, fmt.Errorf("invalid active areas provided")
		}
	}

	userType := user.Type(*data.Type)
	passportIssueDate := time.Time(data.PassportIssueDate)
	defaultRole, err := DefaultUserRole(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("unable to find default role, %w", err)
	}
	createdUser, err = tx.User.
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
		SetPassword(hashedPassword).
		SetRole(defaultRole).
		Save(ctx)
	return
}

func (r *userRepository) ConfirmRegistration(ctx context.Context, login string) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	_, err = tx.User.Update().Where(user.LoginEQ(login)).SetIsRegistrationConfirmed(true).Save(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, userId int) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}

	_, err = tx.User.UpdateOneID(userId).SetIsDeleted(true).Save(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepository) SetIsReadonly(ctx context.Context, id int, isReadonly bool) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	_, err = tx.User.UpdateOneID(id).SetIsReadonly(isReadonly).Save(ctx)
	if err != nil {
		return err
	}
	return nil
}
