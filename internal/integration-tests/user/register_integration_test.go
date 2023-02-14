package user

import (
	"context"
	"net/http"
	"testing"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/users"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/integration-tests/common"

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
		// not provided in documentation, but also required
		// todo: update documentation
		Name: gofakeit.Name(),
		// not provided in documentation, but also required
		// todo: update documentation
		Email: strfmt.Email(gofakeit.Email()),
	}

	params := users.NewPostUserParams()
	params.SetContext(ctx)
	params.SetData(data)
	params.SetHTTPClient(http.DefaultClient)

	t.Run("user person register ok", func(t *testing.T) {
		// login, password, type required, name and email also
		user, err := c.Users.PostUser(params)
		require.NoError(t, err)

		assert.Equal(t, data.Login, user.Payload.Data.Login)
	})

	t.Run("login is already used", func(t *testing.T) {
		r, err := c.Users.PostUser(params)
		require.Error(t, err, r)

		errExp := users.NewPostUserDefault(417)
		errExp.Payload = &models.Error{
			Data: &models.ErrorData{
				CorrelationID: "",
				Message:       "login is already used",
			},
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
			// not provided in documentation, but also required
			// todo: update documentation
			Name: gofakeit.Name(),
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
			// not provided in documentation, but also required
			// todo: update documentation
			Name: gofakeit.Name(),
			// not provided in documentation, but also required
			// todo: update documentation
			Email: strfmt.Email(gofakeit.Email()),
		}
		params = users.NewPostUserParams()
		params.SetData(data)

		_, err = c.Users.PostUser(params)
		require.Error(t, err)

		errExp := users.NewPostUserDefault(422)
		errExp.Payload = &models.Error{
			Data: nil,
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
		errExp.Payload = &models.Error{
			Data: nil,
		}

		assert.Equal(t, errExp, err)
	})

	t.Run("name validation error", func(t *testing.T) {
		empty := ""
		data.Name = empty
		params.SetData(data)

		_, err := c.Users.PostUser(params)
		require.Error(t, err)

		errExp := users.NewPostUserDefault(422)
		errExp.Payload = &models.Error{
			Data: nil,
		}

		assert.Equal(t, errExp, err)
	})

	t.Run("email validation error", func(t *testing.T) {
		empty := ""
		data.Email = strfmt.Email(empty)
		params.SetData(data)

		_, err := c.Users.PostUser(params)
		require.Error(t, err)

		errExp := users.NewPostUserDefault(422)
		errExp.Payload = &models.Error{
			Data: nil,
		}

		assert.Equal(t, errExp, err)
	})

	t.Run("password validation error", func(t *testing.T) {
		data.Password = nil
		params.SetData(data)

		_, err := c.Users.PostUser(params)
		require.Error(t, err)

		errExp := users.NewPostUserDefault(422)
		errExp.Payload = &models.Error{
			Data: nil,
		}

		assert.Equal(t, errExp, err)
	})

	t.Run("password validation error: length less than 3", func(t *testing.T) {
		notEmpty := "22"
		data.Password = &notEmpty
		params.SetData(data)

		_, err := c.Users.PostUser(params)
		require.Error(t, err)

		errExp := users.NewPostUserDefault(422)
		errExp.Payload = &models.Error{
			Data: nil,
		}

		assert.Equal(t, errExp, err)
	})

	t.Run("type validation error: type is not person or organization", func(t *testing.T) {
		notDefinedType := "some type"
		data.Type = &notDefinedType
		params.SetData(data)

		_, err := c.Users.PostUser(params)
		require.Error(t, err)

		errExp := users.NewPostUserDefault(422)
		errExp.Payload = &models.Error{
			Data: nil,
		}

		assert.Equal(t, errExp, err)
	})
}
