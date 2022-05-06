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

type Auth struct {
	Id    int
	Login string
	Role  *Role
}

func (a *Auth) IsAmin() bool {
	if a.Role == nil {
		return false
	}
	return a.Role.Slug == AdminSlug
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
	if !auth.IsAmin() {
		return false, errors.New("User is not admin")
	}
	return true, nil
}
