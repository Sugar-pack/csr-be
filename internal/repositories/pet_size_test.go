package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/swagger/models"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
)

const petSizeEntityName = "pet_size"

func TestPetSizeRepository_Create(t *testing.T) {
	ctx := context.Background()
	client := getClient(t, petSizeEntityName)
	name := "test"
	size := "size"
	defer client.Close()

	repo := NewPetSizeRepository()
	tx, err := client.Tx(ctx)
	require.NoError(t, err)

	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	petSize, err := repo.Create(ctx, models.PetSize{Name: &name, Size: &size})
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	selectedPetSize, err := client.PetSize.Get(ctx, 1)
	require.NoError(t, err)

	require.Equal(t, petSize.ID, selectedPetSize.ID)
	require.Equal(t, petSize.Name, selectedPetSize.Name)
}

func TestPetSizeRepository_GetAll(t *testing.T) {
	ctx := context.Background()
	client := getClient(t, petSizeEntityName)
	name := "test"
	_, err := client.PetSize.Create().SetName(name).Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	repo := NewPetSizeRepository()
	tx, err := client.Tx(ctx)
	require.NoError(t, err)

	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	rows, err := repo.GetAll(ctx)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	require.Equal(t, 1, rows[0].ID)
	require.Equal(t, name, rows[0].Name)
}

func TestPetSizeRepository_Get(t *testing.T) {
	ctx := context.Background()
	client := getClient(t, petSizeEntityName)
	name := "test"
	_, err := client.PetSize.Create().SetName(name).Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	repo := NewPetSizeRepository()
	tx, err := client.Tx(ctx)
	require.NoError(t, err)

	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	row, err := repo.GetByID(ctx, 1)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	require.Equal(t, 1, row.ID)
	require.Equal(t, name, row.Name)
}

func TestPetSizeRepository_Delete(t *testing.T) {
	ctx := context.Background()
	client := getClient(t, petSizeEntityName)
	name := "test"
	_, err := client.PetSize.Create().SetName(name).Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	repo := NewPetSizeRepository()
	tx, err := client.Tx(ctx)
	require.NoError(t, err)

	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	err = repo.Delete(ctx, 1)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	selected, err := client.PetSize.Get(ctx, 1)
	require.ErrorContains(t, err, "ent: pet_size not found")
	require.Nil(t, selected)
}

func TestPetSizeRepository_Update(t *testing.T) {
	ctx := context.Background()
	client := getClient(t, petSizeEntityName)
	name := "test"
	name2 := "test2"
	size := "size"
	size2 := "size2"
	_, err := client.PetSize.Create().SetName(name).SetSize(size).Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	repo := NewPetSizeRepository()
	tx, err := client.Tx(ctx)
	require.NoError(t, err)

	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	_, err = repo.Update(ctx, 1, &models.PetSize{Name: &name2, Size: &size2})
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	selected, err := client.PetSize.Get(ctx, 1)
	require.NoError(t, err)

	require.Equal(t, 1, selected.ID)
	require.Equal(t, name2, selected.Name)
	require.Equal(t, size2, selected.Size)
}
