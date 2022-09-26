package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/generated/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
)

type photoRepositorySuite struct {
	suite.Suite
	ctx        context.Context
	client     *ent.Client
	repository PhotoRepository
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
	newPhoto := models.Photo{}

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	createdPhoto, err := s.repository.CreatePhoto(ctx, newPhoto)
	assert.Error(t, err)
	assert.NoError(t, tx.Rollback())
	assert.Errorf(t, err, "id must not be empty")
	assert.Nil(t, createdPhoto)

	_, err = s.client.Photo.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal()
	}
}

func (s *photoRepositorySuite) TestPhotoRepository_CreatePhoto_EmptyFileName() {
	t := s.T()
	id := "somegenerateduuid"
	fileName := "somegenerateduuid.jpg"
	newPhoto := models.Photo{
		ID: &id,
	}

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	createdPhoto, err := s.repository.CreatePhoto(ctx, newPhoto)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
	assert.Equal(t, id, createdPhoto.ID)
	assert.Equal(t, fileName, createdPhoto.FileName)

	_, err = s.client.Photo.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal()
	}
}

func (s *photoRepositorySuite) TestPhotoRepository_CreatePhoto_OK() {
	t := s.T()
	id := "somegenerateduuid"
	fileName := "somegenerateduuid.jpg"
	newPhoto := models.Photo{
		ID:       &id,
		FileName: fileName,
	}

	ctx := s.ctx
	tx, err := s.client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	createdPhoto, err := s.repository.CreatePhoto(ctx, newPhoto)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
	assert.Equal(t, id, createdPhoto.ID)
	assert.Equal(t, fileName, createdPhoto.FileName)

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
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	photo, err := s.repository.PhotoByID(ctx, id)
	assert.Error(t, err)
	assert.NoError(t, tx.Rollback())
	assert.Errorf(t, err, "not found")
	assert.Nil(t, photo)

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
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	photo, err := s.repository.PhotoByID(ctx, createdPhoto.ID)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
	assert.Equal(t, id, photo.ID)
	assert.Equal(t, fileName, photo.FileName)

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
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	err = s.repository.DeletePhotoByID(ctx, id)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())

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
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	err = s.repository.DeletePhotoByID(ctx, createdPhoto.ID)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())

	_, err = s.client.Photo.Delete().Exec(s.ctx)
	if err != nil {
		t.Fatal()
	}
}
