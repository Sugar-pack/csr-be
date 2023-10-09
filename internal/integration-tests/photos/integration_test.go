package photos

import (
	"context"
	"errors"
	"flag"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/client/photos"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/integration-tests/common"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/messages"
)

var (
	auth runtime.ClientAuthInfoWriterFunc
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Short() {

		t := &testing.T{}
		loginUser := common.AdminUserLogin(t)

		auth = common.AuthInfoFunc(loginUser.GetPayload().AccessToken)

		os.Exit(m.Run())
	}
}

func TestIntegration_PhotosUpload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	beClient := common.SetupClient()

	t.Run("Create New Photo Ok JPEG", func(t *testing.T) {
		fileName := "../common/cat.jpeg"
		f, err := os.Open(fileName)
		require.NoError(t, err)

		params := photos.NewCreateNewPhotoParamsWithContext(ctx).WithFile(f)
		res, err := beClient.Photos.CreateNewPhoto(params, auth)
		require.NoError(t, err)

		require.NotNil(t, res.Payload.Data)
		assert.NotEmpty(t, res.Payload.Data.ID)
		assert.NotEmpty(t, res.Payload.Data.FileName)

		// cleanup
		_, err = beClient.Photos.DeletePhoto(photos.NewDeletePhotoParamsWithContext(ctx).WithPhotoID(*res.Payload.Data.ID), auth)
		require.NoError(t, err)
		f.Close()
	})

	t.Run("Create New Photo Ok JPG", func(t *testing.T) {
		fileName := "../common/cat2.jpg"
		f, err := os.Open(fileName)
		require.NoError(t, err)

		params := photos.NewCreateNewPhotoParamsWithContext(ctx).WithFile(f)
		res, err := beClient.Photos.CreateNewPhoto(params, auth)
		require.NoError(t, err)

		assert.NotEmpty(t, res.Payload.Data.ID)
		assert.NotEmpty(t, res.Payload.Data.FileName)

		// cleanup
		_, err = beClient.Photos.DeletePhoto(photos.NewDeletePhotoParamsWithContext(ctx).WithPhotoID(*res.Payload.Data.ID), auth)
		require.NoError(t, err)
		f.Close()
	})

	t.Run("Create New Photo failed: Wrong file format. File should be jpg or jpeg", func(t *testing.T) {
		fileName := "../common/cat3.png"
		f, err := os.Open(fileName)
		require.NoError(t, err)

		params := photos.NewCreateNewPhotoParamsWithContext(ctx).WithFile(f)
		_, err = beClient.Photos.CreateNewPhoto(params, auth)
		require.Error(t, err)

		var phErr *photos.CreateNewPhotoBadRequest
		require.True(t, errors.As(err, &phErr))

		wantMessage := "Wrong file format. File should be jpg or jpeg"
		require.NotNil(t, phErr.Payload.Message)
		assert.Equal(t, wantMessage, *phErr.Payload.Message)
		f.Close()
	})

	t.Run("Create New Photo failed: empty file - swagger 'multipart: NextPart: EOF'", func(t *testing.T) {
		emptyFile, err := os.Create("empty.jpeg")
		require.NoError(t, err)

		params := photos.NewCreateNewPhotoParamsWithContext(ctx).WithFile(emptyFile)
		_, err = beClient.Photos.CreateNewPhoto(params, auth)
		require.Error(t, err)

		wantErr := photos.NewCreateNewPhotoBadRequest()
		msgExp := "multipart: NextPart: EOF"
		codeExp := int32(http.StatusBadRequest)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &msgExp,
		}
		assert.Equal(t, wantErr, err)

		emptyFile.Close()
		require.NoError(t, os.Remove("empty.jpeg"))
	})

	t.Run("Create New Photo failed: access failed", func(t *testing.T) {
		fileName := "../common/cat2.jpg"
		f, err := os.Open(fileName)
		require.NoError(t, err)

		token := common.TokenNotExist
		_, gotErr := beClient.Photos.CreateNewPhoto(photos.NewCreateNewPhotoParamsWithContext(ctx).WithFile(f),
			common.AuthInfoFunc(&token))
		require.Error(t, gotErr)

		wantErr := photos.NewCreateNewPhotoDefault(http.StatusUnauthorized)
		codeExp := int32(http.StatusUnauthorized)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrInvalidToken,
		}
		assert.Equal(t, wantErr, gotErr)
		f.Close()
	})
}

func TestIntegration_DeletePhoto(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	beClient := common.SetupClient()

	fileName := "../common/cat.jpeg"
	f, err := os.Open(fileName)
	require.NoError(t, err)

	createParams := photos.NewCreateNewPhotoParamsWithContext(ctx).WithFile(f)
	res, err := beClient.Photos.CreateNewPhoto(createParams, auth)
	require.NoError(t, err)

	t.Run("Delete Photo failed: access failed", func(t *testing.T) {
		token := common.TokenNotExist
		_, gotErr := beClient.Photos.DeletePhoto(photos.NewDeletePhotoParamsWithContext(ctx).WithPhotoID(*res.Payload.Data.ID),
			common.AuthInfoFunc(&token))

		require.Error(t, gotErr)

		wantErr := photos.NewDeletePhotoDefault(http.StatusUnauthorized)
		codeExp := int32(http.StatusUnauthorized)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrInvalidToken,
		}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Delete Photo failed: swagger validation, photoID not provided", func(t *testing.T) {
		_, gotErr := beClient.Photos.DeletePhoto(photos.NewDeletePhotoParamsWithContext(ctx), auth)
		require.Error(t, gotErr)

		wantErr := photos.NewDeletePhotoDefault(http.StatusUnprocessableEntity)
		msgExp := "equipmentId in path must be of type int64: \"photos\""
		codeExp := int32(601)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &msgExp,
		}
		assert.Equal(t, wantErr, gotErr)
	})

	t.Run("Delete Photo Ok", func(t *testing.T) {
		result, err := beClient.Photos.DeletePhoto(photos.NewDeletePhotoParamsWithContext(ctx).WithPhotoID(*res.Payload.Data.ID), auth)
		require.NoError(t, err)

		got := result.Payload
		want := "photo deleted"
		assert.Equal(t, want, got)
	})

	t.Run("Delete Photo failed: trying to delete again the same photo, photo not found", func(t *testing.T) {
		_, gotErr := beClient.Photos.DeletePhoto(photos.NewDeletePhotoParamsWithContext(ctx).WithPhotoID(*res.Payload.Data.ID), auth)
		require.Error(t, gotErr)

		wantErr := photos.NewDeletePhotoDefault(http.StatusInternalServerError)
		codeExp := int32(http.StatusInternalServerError)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrDeletePhoto,
			Details: "ent: photo not found",
		}
		assert.Equal(t, wantErr, gotErr)
	})

	f.Close()
}

func TestIntegration_PhotosDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	beClient := common.SetupClient()

	t.Run("Download Photo Ok JPEG", func(t *testing.T) {
		fileName := "../common/cat.jpeg"
		f, err := os.Open(fileName)
		require.NoError(t, err)

		createParams := photos.NewCreateNewPhotoParamsWithContext(ctx).WithFile(f)
		res, err := beClient.Photos.CreateNewPhoto(createParams, auth)
		require.NoError(t, err)

		result, err := beClient.Photos.DownloadPhoto(photos.NewDownloadPhotoParamsWithContext(ctx).WithPhotoID(*res.Payload.Data.ID), auth, io.Discard)
		require.NoError(t, err)

		assert.Equal(t, io.Discard, result.Payload)
		// cleanup
		_, err = beClient.Photos.DeletePhoto(photos.NewDeletePhotoParamsWithContext(ctx).WithPhotoID(*res.Payload.Data.ID), auth)
		require.NoError(t, err)
		f.Close()
	})

	t.Run("Download Photo Ok JPG", func(t *testing.T) {
		fileName := "../common/cat2.jpg"
		f, err := os.Open(fileName)
		require.NoError(t, err)

		createParams := photos.NewCreateNewPhotoParamsWithContext(ctx).WithFile(f)
		res, err := beClient.Photos.CreateNewPhoto(createParams, auth)
		require.NoError(t, err)

		result, err := beClient.Photos.DownloadPhoto(photos.NewDownloadPhotoParamsWithContext(ctx).WithPhotoID(*res.Payload.Data.ID), auth, io.Discard)
		require.NoError(t, err)

		assert.Equal(t, io.Discard, result.Payload)
		// cleanup
		_, err = beClient.Photos.DeletePhoto(photos.NewDeletePhotoParamsWithContext(ctx).WithPhotoID(*res.Payload.Data.ID), auth)
		require.NoError(t, err)
		f.Close()
	})

	t.Run("Download Photo failed: access failed", func(t *testing.T) {
		fileName := "../common/cat.jpeg"
		f, err := os.Open(fileName)
		require.NoError(t, err)

		createParams := photos.NewCreateNewPhotoParamsWithContext(ctx).WithFile(f)
		res, err := beClient.Photos.CreateNewPhoto(createParams, auth)
		require.NoError(t, err)

		token := common.TokenNotExist
		_, gotErr := beClient.Photos.DownloadPhoto(photos.NewDownloadPhotoParamsWithContext(ctx).WithPhotoID(*res.Payload.Data.ID),
			common.AuthInfoFunc(&token), io.Discard)
		require.Error(t, gotErr)

		wantErr := photos.NewDownloadPhotoDefault(http.StatusUnauthorized)
		codeExp := int32(http.StatusUnauthorized)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrInvalidToken,
		}
		assert.Equal(t, wantErr, gotErr)
		f.Close()
	})

	t.Run("Download Photo failed: photoID not provided", func(t *testing.T) {
		_, gotErr := beClient.Photos.DownloadPhoto(photos.NewDownloadPhotoParamsWithContext(ctx), auth, io.Discard)
		require.Error(t, gotErr)

		wantErr := photos.NewDownloadPhotoDefault(http.StatusInternalServerError)
		msgExp := "failed to get photo"
		codeExp := int32(http.StatusInternalServerError)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &msgExp,
			Details: "ent: photo not found",
		}
		assert.Equal(t, wantErr, gotErr)
	})
}

func TestIntegration_PhotoGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	beClient := common.SetupClient()

	t.Run("Get Photo Ok JPEG", func(t *testing.T) {
		fileName := "../common/cat.jpeg"
		f, err := os.Open(fileName)
		require.NoError(t, err)

		createParams := photos.NewCreateNewPhotoParamsWithContext(ctx).WithFile(f)
		res, err := beClient.Photos.CreateNewPhoto(createParams, auth)
		require.NoError(t, err)

		result, err := beClient.Photos.GetPhoto(photos.NewGetPhotoParamsWithContext(ctx).WithPhotoID(*res.Payload.Data.ID), auth, io.Discard)
		require.NoError(t, err)

		assert.Equal(t, io.Discard, result.Payload)
		// cleanup
		_, err = beClient.Photos.DeletePhoto(photos.NewDeletePhotoParamsWithContext(ctx).WithPhotoID(*res.Payload.Data.ID), auth)
		require.NoError(t, err)
		f.Close()
	})

	t.Run("Get Photo Ok JPG", func(t *testing.T) {
		fileName := "../common/cat2.jpg"
		f, err := os.Open(fileName)
		require.NoError(t, err)

		createParams := photos.NewCreateNewPhotoParamsWithContext(ctx).WithFile(f)
		res, err := beClient.Photos.CreateNewPhoto(createParams, auth)
		require.NoError(t, err)

		result, err := beClient.Photos.GetPhoto(photos.NewGetPhotoParamsWithContext(ctx).WithPhotoID(*res.Payload.Data.ID), auth, io.Discard)
		require.NoError(t, err)

		assert.Equal(t, io.Discard, result.Payload)
		// cleanup
		_, err = beClient.Photos.DeletePhoto(photos.NewDeletePhotoParamsWithContext(ctx).WithPhotoID(*res.Payload.Data.ID), auth)
		require.NoError(t, err)
		f.Close()
	})

	t.Run("Get Photo failed: access failed", func(t *testing.T) {
		fileName := "../common/cat.jpeg"
		f, err := os.Open(fileName)
		require.NoError(t, err)

		createParams := photos.NewCreateNewPhotoParamsWithContext(ctx).WithFile(f)
		res, err := beClient.Photos.CreateNewPhoto(createParams, auth)
		require.NoError(t, err)
		token := common.TokenNotExist
		_, gotErr := beClient.Photos.GetPhoto(photos.NewGetPhotoParamsWithContext(ctx).WithPhotoID(*res.Payload.Data.ID), common.AuthInfoFunc(&token), io.Discard)
		require.Error(t, gotErr)

		wantErr := photos.NewGetPhotoDefault(http.StatusUnauthorized)
		codeExp := int32(http.StatusUnauthorized)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &messages.ErrInvalidToken,
		}
		assert.Equal(t, wantErr, gotErr)
		f.Close()
	})

	t.Run("Get Photo failed: swagger validation, photoID not provided", func(t *testing.T) {
		_, gotErr := beClient.Photos.GetPhoto(photos.NewGetPhotoParamsWithContext(ctx), auth, io.Discard)
		require.Error(t, gotErr)

		wantErr := photos.NewGetPhotoDefault(http.StatusUnprocessableEntity)
		msgExp := "equipmentId in path must be of type int64: \"photos\""
		codeExp := int32(601)
		wantErr.Payload = &models.SwaggerError{
			Code:    &codeExp,
			Message: &msgExp,
		}
		assert.Equal(t, wantErr, gotErr)
	})
}
