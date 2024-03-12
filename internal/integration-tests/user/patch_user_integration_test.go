package user

import (
	"context"
	"net/http"
	"testing"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/users"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	utils "git.epam.com/epm-lstr/epm-lstr-lc/be/internal/integration-tests/common"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/messages"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_PatchUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	client := utils.SetupClient()

	loginUser := utils.AdminUserLogin(t)
	token := loginUser.GetPayload().AccessToken

	t.Run("patch update successfully passed", func(t *testing.T) {
		var date = time.Date(2009, time.November, 10, 00, 0, 0, 0, time.UTC)
		params := users.NewPatchUserParamsWithContext(ctx)
		name := gofakeit.Name()
		orgName := gofakeit.Company()
		passportAuthority := gofakeit.Name()
		passportIssueDate := date
		passportNumber := gofakeit.Digit()
		passportSeries := gofakeit.Digit()
		phone := gofakeit.Phone()
		surname := gofakeit.LastName()
		vk := gofakeit.URL()
		website := gofakeit.URL()
		params.UserPatch = &models.PatchUserRequest{
			Name:              name,
			OrgName:           orgName,
			PassportAuthority: passportAuthority,
			PassportIssueDate: strfmt.Date(date),
			PassportNumber:    passportNumber,
			PassportSeries:    passportSeries,
			Phone:             phone,
			Surname:           surname,
			Vk:                vk,
			Website:           website,
		}
		_, err := client.Users.PatchUser(params, utils.AuthInfoFunc(token))
		require.NoError(t, err)

		user, err := utils.GetUser(ctx, client, utils.AuthInfoFunc(token))
		require.NoError(t, err)

		assert.Equal(t, name, *user.GetPayload().Name)
		assert.Equal(t, orgName, *user.GetPayload().OrgName)
		assert.Equal(t, passportAuthority, *user.GetPayload().PassportAuthority)
		assert.Equal(t, passportIssueDate.String(), *user.GetPayload().PassportIssueDate)
		assert.Equal(t, passportNumber, *user.GetPayload().PassportNumber)
		assert.Equal(t, passportSeries, *user.GetPayload().PassportSeries)
		assert.Equal(t, phone, *user.GetPayload().PhoneNumber)
		assert.Equal(t, surname, *user.GetPayload().Surname)
	})

	t.Run("patch failed: no authorization", func(t *testing.T) {
		params := users.NewPatchUserParamsWithContext(ctx)
		name := gofakeit.Name()
		params.UserPatch = &models.PatchUserRequest{
			Name: name,
		}
		_, err := client.Users.PatchUser(params, utils.AuthInfoFunc(nil))
		assert.Error(t, err)

		errExp := users.NewPatchUserDefault(http.StatusUnauthorized)
		msgExp := "unauthenticated for invalid credentials"
		codeExp := int32(http.StatusUnauthorized)
		errExp.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &msgExp,
		}
		assert.Equal(t, errExp, err)
	})

	t.Run("patch failed: token contains an invalid number of segments", func(t *testing.T) {
		params := users.NewPatchUserParamsWithContext(ctx)
		name := gofakeit.Name()
		params.UserPatch = &models.PatchUserRequest{
			Name: name,
		}
		dummyToken := utils.TokenNotExist
		_, err := client.Users.PatchUser(params, utils.AuthInfoFunc(&dummyToken))
		assert.Error(t, err)

		errExp := users.NewPatchUserDefault(http.StatusUnauthorized)
		codeExp := int32(http.StatusUnauthorized)
		errExp.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrInvalidToken,
		}
		assert.Equal(t, errExp, err)
	})

	t.Run("patch failed: validation required", func(t *testing.T) {
		params := users.NewPatchUserParamsWithContext(ctx)
		params.UserPatch = nil
		_, err := client.Users.PatchUser(params, utils.AuthInfoFunc(token))
		assert.Error(t, err)

		errExp := users.NewPatchUserDefault(http.StatusUnprocessableEntity)
		msgExp := "userPatch in body is required"
		codeExp := int32(602)
		errExp.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &msgExp,
		}
		assert.Equal(t, errExp, err)
	})

	t.Run("user not updated: empty patch fields", func(t *testing.T) {
		userBeforeUpdate, err := utils.GetUser(ctx, client, utils.AuthInfoFunc(token))
		require.NoError(t, err)

		params := users.NewPatchUserParamsWithContext(ctx)
		params.UserPatch = &models.PatchUserRequest{
			Name:              "",
			OrgName:           "",
			PassportAuthority: "",
			PassportIssueDate: strfmt.Date{},
			PassportNumber:    "",
			PassportSeries:    "",
			Phone:             "",
			Surname:           "",
			Vk:                "",
			Website:           "",
		}
		_, err = client.Users.PatchUser(params, utils.AuthInfoFunc(token))
		assert.NoError(t, err)

		// wait for update apply
		time.Sleep(5 * time.Second)

		userAfterUpdate, err := utils.GetUser(ctx, client, utils.AuthInfoFunc(token))
		require.NoError(t, err)

		assert.Equal(t, userBeforeUpdate, userAfterUpdate)
	})
}
