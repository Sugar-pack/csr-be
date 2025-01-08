package user

import (
	"context"
	"net/http"
	"testing"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/users"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/integration-tests/common"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/messages"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_RegisterUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	c := utils.SetupClient()

	l, p, err := utils.GenerateLoginAndPassword()
	require.NoError(t, err)

	userType := "person"
	data := &models.UserRegister{
		ActiveAreas: []int64{1},
		Login:       &l,
		Password:    &p,
		Type:        &userType,
		Name:        gofakeit.Name(),
		// not provided in documentation, but also required
		// todo: update documentation
		Email: strfmt.Email(gofakeit.Email()),
	}

	params := users.NewPostUserParams()
	params.SetContext(ctx)
	params.SetData(data)
	params.SetHTTPClient(http.DefaultClient)

	t.Run("user person register ok", func(t *testing.T) {
		// login, password, type required
		user, err := c.Users.PostUser(params)
		require.NoError(t, err)

		assert.Equal(t, data.Login, user.Payload.Data.Login)
	})

	t.Run("login is already used", func(t *testing.T) {
		r, err := c.Users.PostUser(params)
		require.Error(t, err, r)

		errExp := users.NewPostUserDefault(http.StatusExpectationFailed)
		codeExp := int32(http.StatusExpectationFailed)
		errExp.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrLoginInUse,
		}
		assert.Equal(t, errExp, err)
	})

	t.Run("user organization register ok", func(t *testing.T) {
		// login, password, type required, name and email also
		userType = "organization"
		l, p, err := utils.GenerateLoginAndPassword()
		require.NoError(t, err)

		data = &models.UserRegister{
			ActiveAreas: []int64{2},
			Login:       &l,
			Password:    &p,
			Type:        &userType,
			Name:        gofakeit.Name(),
			// not provided in documentation, but also required
			// todo: update documentation
			Email: strfmt.Email(gofakeit.Email()),
		}
		params = users.NewPostUserParams()
		params.SetData(data)

		user, err := c.Users.PostUser(params)
		require.NoError(t, err)

		assert.Equal(t, data.Login, user.Payload.Data.Login)
	})

	t.Run("user register failed: type not implemented", func(t *testing.T) {
		// login, password, type required, name and email also
		userType = "dummy type"
		l, p, err := utils.GenerateLoginAndPassword()
		require.NoError(t, err)

		data = &models.UserRegister{
			ActiveAreas: []int64{3},
			Login:       &l,
			Password:    &p,
			Type:        &userType,
			Name:        gofakeit.Name(),
			// not provided in documentation, but also required
			// todo: update documentation
			Email: strfmt.Email(gofakeit.Email()),
		}
		params = users.NewPostUserParams()
		params.SetData(data)

		_, err = c.Users.PostUser(params)
		require.Error(t, err)

		errExp := users.NewPostUserDefault(422)
		msgExp := "type in body should be one of [person organization]"
		codeExp := int32(606)
		errExp.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &msgExp,
		}

		assert.Equal(t, errExp, err)
	})

	t.Run("login validation error", func(t *testing.T) {
		empty := ""
		data.Login = &empty
		params.SetData(data)

		_, err := c.Users.PostUser(params)
		require.Error(t, err)

		errExp := users.NewPostUserDefault(422)
		msgExp := "login in body should be at least 3 chars long"
		codeExp := int32(604)
		errExp.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &msgExp,
		}

		assert.Equal(t, errExp, err)
	})

	t.Run("name absence is ok", func(t *testing.T) {
		userType = "person"
		l, p, err = utils.GenerateLoginAndPassword()
		require.NoError(t, err)
		data = &models.UserRegister{
			ActiveAreas: []int64{3},
			Login:       &l,
			Password:    &p,
			Type:        &userType,
			Name:        "d",
			Email:       strfmt.Email(gofakeit.Email()),
		}
		params.SetData(data)
		_, err := c.Users.PostUser(params)
		require.NoError(t, err)
	})

	t.Run("email validation error", func(t *testing.T) {
		userType = "person"
		l, p, err = utils.GenerateLoginAndPassword()
		require.NoError(t, err)
		data = &models.UserRegister{
			ActiveAreas: []int64{3},
			Login:       &l,
			Password:    &p,
			Type:        &userType,
			Name:        gofakeit.Name(),
			Email:       strfmt.Email(""),
		}
		params.SetData(data)
		_, err := c.Users.PostUser(params)
		require.Error(t, err)

		errExp := users.NewPostUserDefault(http.StatusInternalServerError)
		codeExp := int32(http.StatusInternalServerError)
		errExp.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrCreateUser,
			Details: "ent: validator failed for field \"User.email\": value is less than the required length",
		}
		assert.Equal(t, errExp, err)
	})

	t.Run("password validation error", func(t *testing.T) {
		data = &models.UserRegister{
			ActiveAreas: []int64{3},
			Login:       &l,
			Password:    nil,
			Type:        &userType,
			Name:        gofakeit.Name(),
			Email:       strfmt.Email(gofakeit.Email()),
		}
		params.SetData(data)

		_, err := c.Users.PostUser(params)
		require.Error(t, err)

		errExp := users.NewPostUserDefault(422)
		msgExp := "password in body is required"
		codeExp := int32(602)
		errExp.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &msgExp,
		}
		assert.Equal(t, errExp, err)
	})

	t.Run("password validation error: length less than 3", func(t *testing.T) {
		data = &models.UserRegister{
			ActiveAreas: []int64{3},
			Login:       &l,
			Password:    &p,
			Type:        &userType,
			Name:        gofakeit.Name(),
			Email:       strfmt.Email(gofakeit.Email()),
		}

		notEmpty := "22"
		data.Password = &notEmpty
		params.SetData(data)

		_, err := c.Users.PostUser(params)
		require.Error(t, err)

		errExp := users.NewPostUserDefault(422)
		msgExp := "password in body should be at least 6 chars long"
		codeExp := int32(604)
		errExp.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &msgExp,
		}

		assert.Equal(t, errExp, err)
	})

	t.Run("type validation error: type is not person or organization", func(t *testing.T) {
		data = &models.UserRegister{
			ActiveAreas: []int64{3},
			Login:       &l,
			Password:    &p,
			Type:        &userType,
			Name:        gofakeit.Name(),
			// not provided in documentation, but also required
			// todo: update documentation
			Email: strfmt.Email(gofakeit.Email()),
		}

		notDefinedType := "some type"
		data.Type = &notDefinedType
		params.SetData(data)

		_, err := c.Users.PostUser(params)
		require.Error(t, err)

		errExp := users.NewPostUserDefault(422)
		msgExp := "type in body should be one of [person organization]"
		codeExp := int32(606)
		errExp.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &msgExp,
		}

		assert.Equal(t, errExp, err)
	})
}
