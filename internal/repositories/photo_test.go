package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

type photoRepositorySuite struct {
	suite.Suite
	ctx        context.Context
	client     *ent.Client
	repository domain.PhotoRepository
}

func TestPhotoSuite(t *testing.T) {
	suite.Run(t, new(photoRepositorySuite))
}

func (s *photoRepositorySuite) SetupTest() {
	t := s.T()
	s.ctx = context.Background()
	client := enttest.Open(t, "sqlite3", "file:photo?mode=memory&cache=shared&_fk=1")
	s.client = client
	s.repository = NewPhotoRepository()
}

func (s *photoRepositorySuite) TearDownSuite() {
	s.client.Close()
}

func (s *photoRepositorySuite) TestPhotoRepository_CreatePhoto_EmptyID() {
	t := s.T()
	newPhoto := &ent.Photo{}

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	createdPhoto, err := s.repository.CreatePhoto(ctx, newPhoto)
	require.Error(t, err)
	require.NoError(t, tx.Rollback())
	require.Errorf(t, err, "id must not be empty")
	require.Nil(t, createdPhoto)

	_, err = s.client.Photo.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal()
	}
}

func (s *photoRepositorySuite) TestPhotoRepository_CreatePhoto_EmptyFileName() {
	t := s.T()
	id := "somegenerateduuid"
	fileName := "somegenerateduuid.jpg"
	newPhoto := &ent.Photo{
		ID: id,
	}

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	createdPhoto, err := s.repository.CreatePhoto(ctx, newPhoto)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, id, createdPhoto.ID)
	require.Equal(t, fileName, createdPhoto.FileName)

	_, err = s.client.Photo.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal()
	}
}

func (s *photoRepositorySuite) TestPhotoRepository_CreatePhoto_OK() {
	t := s.T()
	id := "somegenerateduuid"
	fileName := "somegenerateduuid.jpg"
	newPhoto := &ent.Photo{
		ID:       id,
		FileName: fileName,
	}

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	createdPhoto, err := s.repository.CreatePhoto(ctx, newPhoto)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, id, createdPhoto.ID)
	require.Equal(t, fileName, createdPhoto.FileName)

	_, err = s.client.Photo.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal()
	}
}

func (s *photoRepositorySuite) TestPhotoRepository_PhotoByID_NotFound() {
	t := s.T()
	id := "somegenerateduuid"

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	photo, err := s.repository.PhotoByID(ctx, id)
	require.Error(t, err)
	require.NoError(t, tx.Rollback())
	require.Errorf(t, err, "not found")
	require.Nil(t, photo)

	_, err = s.client.Photo.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal()
	}
}

func (s *photoRepositorySuite) TestPhotoRepository_PhotoByID_OK() {
	t := s.T()
	id := "somegenerateduuid"
	fileName := "somegenerateduuid.jpg"
	createdPhoto, err := s.client.Photo.Create().SetID(id).
		SetFileName(fileName).
		Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	photo, err := s.repository.PhotoByID(ctx, createdPhoto.ID)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	require.Equal(t, id, photo.ID)
	require.Equal(t, fileName, photo.FileName)

	_, err = s.client.Photo.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal()
	}
}

func (s *photoRepositorySuite) TestPhotoRepository_DeletePhotoByID_NotExistsOK() {
	t := s.T()
	id := "somegenerateduuid"

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	err = s.repository.DeletePhotoByID(ctx, id)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	_, err = s.client.Photo.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal()
	}
}

func (s *photoRepositorySuite) TestPhotoRepository_DeletePhotoByID_OK() {
	t := s.T()
	id := "somegenerateduuid"
	fileName := "somegenerateduuid.jpg"
	createdPhoto, err := s.client.Photo.Create().SetID(id).
		SetFileName(fileName).
		Save(s.ctx)
	if err != nil {
		t.Fatal(err)
	}

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	require.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	err = s.repository.DeletePhotoByID(ctx, createdPhoto.ID)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	_, err = s.client.Photo.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal()
	}
}
