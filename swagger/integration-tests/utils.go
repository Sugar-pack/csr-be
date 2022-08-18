package integration_tests

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/kinds"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/pet_kind"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/pet_size"
	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/client"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/client/users"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/config"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

var (
	Login    string
	Password string
	Name     string
	Email    string
)

var UserID int64

const (
	LoginNotExist    = "some dummy login"
	PasswordNotExist = "some dummy password"
	TokenNotExist    = "some dummy token"
)

func GenerateLoginAndPassword() (string, string, error) {
	login := gofakeit.Username()
	generator, err := utils.NewPasswordGenerator(6)
	if err != nil {
		return "", "", err
	}
	password, err := generator.NewPassword()
	if err != nil {
		return "", "", err
	}
	return login, password, nil
}

func CreateUser(ctx context.Context, c *client.Be, login, password string) (*models.CreateUserResponseData, error) {
	userType := "person"
	Name = gofakeit.Name()
	Email = gofakeit.Email()
	data := &models.UserRegister{
		ActiveAreas: []int64{1},
		Login:       &login,
		Password:    &password,
		Type:        &userType,
		// not provided in documentation, but also required
		// todo: update documentation
		Name: Name,
		// not provided in documentation, but also required
		// todo: update documentation
		Email: strfmt.Email(Email),
	}

	params := users.NewPostUserParams()
	params.SetContext(ctx)
	params.SetData(data)
	params.SetHTTPClient(http.DefaultClient)

	user, err := c.Users.PostUser(params)
	if err != nil {
		return nil, err
	}

	payload := user.GetPayload()
	UserID = *payload.Data.ID // 8
	return payload.Data, nil
}

func LoginUser(ctx context.Context, client *client.Be, login, password string) (*users.LoginOK, error) {
	info := &models.LoginInfo{
		Login:    &login,
		Password: &password,
	}
	params := users.NewLoginParams()
	params.SetLogin(info)
	params.SetContext(ctx)
	params.SetHTTPClient(http.DefaultClient)

	return client.Users.Login(params)
}

func GetUser(ctx context.Context, client *client.Be, authInfo runtime.ClientAuthInfoWriter) (*users.GetCurrentUserOK, error) {
	params := users.NewGetCurrentUserParamsWithContext(ctx)

	currentUser, err := client.Users.GetCurrentUser(params, authInfo)
	if err != nil {
		return nil, err
	}
	return currentUser, nil
}

func SetupClient() *client.Be {
	strReg := strfmt.Default
	serverConfig, err := config.SetupServerConfig()
	if err != nil {
		log.Fatal("fail to setup server config", zap.Error(err))
	}
	transportConf := client.DefaultTransportConfig()
	host := "localhost"
	transportConf.Host = fmt.Sprintf("%s:%v", host, serverConfig.Port)
	transportConf.Schemes = []string{"http"}
	swaggerClient := client.NewHTTPClientWithConfig(strReg, transportConf)
	return swaggerClient
}

func AuthInfoFunc(token *string) runtime.ClientAuthInfoWriterFunc {
	authFunc := runtime.ClientAuthInfoWriterFunc(func(r runtime.ClientRequest, registry strfmt.Registry) error {
		// todo: get rid of time.Sleep(), it's temporary added to avoid flaky error 500
		// "ERROR: duplicate key value violates unique constraint \"tokens_access_token_key\" (SQLSTATE 23505)"
		time.Sleep(time.Second)
		if token == nil {
			return nil
		}
		return r.SetHeaderParam(runtime.HeaderAuthorization, *token)
	})
	return authFunc
}

func CreatePetSize(ctx context.Context, client *client.Be, token, size, name string) (*pet_size.CreateNewPetSizeCreated, error) {
	params := pet_size.NewCreateNewPetSizeParamsWithContext(ctx)
	params.NewPetSize = &models.PetSize{
		// should not provide ID
		Name: &name,
		Size: &size,
	}
	return client.PetSize.CreateNewPetSize(params, AuthInfoFunc(&token))
}

func CreatePetKind(ctx context.Context, client *client.Be, token, name string) (*pet_kind.CreateNewPetKindCreated, error) {
	params := pet_kind.NewCreateNewPetKindParamsWithContext(ctx)
	params.NewPetKind = &models.PetKind{
		// should not provide ID
		Name: &name,
	}
	return client.PetKind.CreateNewPetKind(params, AuthInfoFunc(&token))
}

func CreateKind(ctx context.Context, client *client.Be, token, name string, maxReservationTime, maxReservationUnits int64) (*kinds.CreateNewKindCreated, error) {
	params := kinds.NewCreateNewKindParamsWithContext(ctx)
	params.NewKind = &models.CreateNewKind{
		MaxReservationTime:  &maxReservationTime,
		MaxReservationUnits: &maxReservationUnits,
		Name:                &name,
	}
	return client.Kinds.CreateNewKind(params, AuthInfoFunc(&token))
}
