package common

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/config"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/users"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/utils"
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
	generator, err := utils.NewPasswordGenerator(8)
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

func AdminLoginPassword(t *testing.T) (string, string, int64) {
	t.Helper()
	l, p, err := GenerateLoginAndPassword()
	require.NoError(t, err)

	ctx := context.Background()
	client := SetupClient()

	user, err := CreateUser(ctx, client, l, p)
	require.NoError(t, err)

	// login and get token
	loginUser, err := LoginUser(ctx, client, l, p)
	require.NoError(t, err)
	auth := AuthInfoFunc(loginUser.GetPayload().AccessToken)

	role := int64(1) //TODO: use const
	params := &users.AssignRoleToUserParams{
		UserID: *user.ID,
		Data: &models.AssignRoleToUser{
			RoleID: &role,
		},
	}
	params.SetContext(ctx)
	params.SetHTTPClient(http.DefaultClient)

	_, err = client.Users.AssignRoleToUser(params, auth)
	require.NoError(t, err)
	return l, p, *user.ID
}

func AdminUserLogin(t *testing.T) *users.LoginOK {
	t.Helper()
	ctx := context.Background()
	client := SetupClient()
	l, p, _ := AdminLoginPassword(t)
	// login and get token with admin role
	loginUser, err := LoginUser(ctx, client, l, p)
	require.NoError(t, err)
	return loginUser
}

func SetupClient() *client.Be {
	serverConfig, err := config.GetAppConfig("../../../int-test-infra/")
	if err != nil {
		log.Fatal("fail to setup server config", zap.Error(err))
	}

	host := "localhost"
	schemes := []string{"http"}
	apiURL := fmt.Sprintf("%s:%v", host, serverConfig.Server.Port)

	swaggerClient, err := NewAPIClient(apiURL, schemes)
	if err != nil {
		log.Fatal("fail to setup client", zap.Error(err))
	}

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

func NewAPIClient(host string, schemes []string) (*client.Be, error) {
	be := httptransport.New(host, client.DefaultBasePath, schemes)
	// Generated client does not accept content types specified in schema
	// https://github.com/go-swagger/go-swagger/issues/1244
	be.Consumers["image/jpg"] = runtime.ByteStreamConsumer()
	return client.New(be, nil), nil
}
