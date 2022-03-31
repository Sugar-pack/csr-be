package authentication

import (
	"errors"
)

type Role struct {
	Id   int
	Slug string
}

type Auth struct {
	Login string
	Role  *Role
}

func (a *Auth) IsAmin() bool {
	if a.Role == nil {
		return false
	}
	return a.Role.Slug == "administrator"
}

func GetAuth(access interface{}) (*Auth, error) {
	auth, ok := access.(Auth)
	if !ok {
		return nil, errors.New("unable to convert to Auth type")
	}
	return &auth, nil
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
