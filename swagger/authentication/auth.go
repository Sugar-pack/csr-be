package authentication

import (
	"errors"
	"fmt"
)

type Role struct {
	Id   int
	Slug string
}

const AdminSlug = "administrator"
const ManagerSlug = "manager"
const OperatorSlug = "operator"

type Auth struct {
	Id    int
	Login string
	Role  *Role
}

func (a *Auth) IsAdmin() bool {
	if a.Role == nil {
		return false
	}
	return a.Role.Slug == AdminSlug
}

func (a *Auth) IsManager() bool {
	if a.Role == nil {
		return false
	}
	return a.Role.Slug == ManagerSlug
}

func (a *Auth) IsOperator() bool {
	if a.Role == nil {
		return false
	}
	return a.Role.Slug == OperatorSlug
}

func GetAuth(access interface{}) (*Auth, error) {
	auth, ok := access.(Auth)
	if !ok {
		return nil, errors.New("unable to convert to Auth type")
	}
	return &auth, nil
}

func GetUserId(access interface{}) (int, error) {
	auth, err := GetAuth(access)
	if err != nil {
		return 0, fmt.Errorf("get user id error, failed to get auth: %s", err)
	}
	return auth.Id, nil
}

func IsAdmin(access interface{}) (bool, error) {
	auth, err := GetAuth(access)
	if err != nil {
		return false, err
	}
	if !auth.IsAdmin() {
		return false, errors.New("user is not admin")
	}
	return true, nil
}

func IsManager(access interface{}) (bool, error) {
	auth, err := GetAuth(access)
	if err != nil {
		return false, err
	}
	if !auth.IsManager() {
		return false, errors.New("user is not manager")
	}
	return true, nil
}

func IsOperator(access interface{}) (bool, error) {
	auth, err := GetAuth(access)
	if err != nil {
		return false, err
	}
	if !auth.IsOperator() {
		return false, errors.New("user is not operator")
	}
	return true, nil
}
