package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
)

const equipmentStatusNameEntityName = "equipment_status_name"

func TestEquipmentStatusNameRepository_Create(t *testing.T) {
	ctx := context.Background()
	client := getClient(t, equipmentStatusNameEntityName)
	statusName := "test"

	defer client.Close()

	repo := NewEquipmentStatusNameRepository()
	tx, err := client.Tx(ctx)
	require.NoError(t, err)

	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	status, err := repo.Create(ctx, statusName)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	selectedStatus, err := client.EquipmentStatusName.Get(ctx, 1)
	require.NoError(t, err)

	require.Equal(t, status.ID, selectedStatus.ID)
	require.Equal(t, status.Name, selectedStatus.Name)
}

func TestEquipmentStatusNameRepository_GetAll(t *testing.T) {
	ctx := context.Background()
	client := getClient(t, equipmentStatusNameEntityName)
	statusName := "test"
	_, err := client.EquipmentStatusName.Create().SetName(statusName).Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	repo := NewEquipmentStatusNameRepository()
	tx, err := client.Tx(ctx)
	require.NoError(t, err)

	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	statuses, err := repo.GetAll(ctx)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	require.Equal(t, 1, statuses[0].ID)
	require.Equal(t, statusName, statuses[0].Name)
}

func TestEquipmentStatusNameRepository_Get(t *testing.T) {
	ctx := context.Background()
	client := getClient(t, equipmentStatusNameEntityName)
	statusName := "test"
	_, err := client.EquipmentStatusName.Create().SetName(statusName).Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	repo := NewEquipmentStatusNameRepository()
	tx, err := client.Tx(ctx)
	require.NoError(t, err)

	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	status, err := repo.Get(ctx, 1)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	require.Equal(t, 1, status.ID)
	require.Equal(t, statusName, status.Name)
}
func TestEquipmentStatusNameRepository_GetByName(t *testing.T) {
	ctx := context.Background()
	client := getClient(t, equipmentStatusNameEntityName)
	statusName := "test"
	_, err := client.EquipmentStatusName.Create().SetName(statusName).Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	repo := NewEquipmentStatusNameRepository()
	tx, err := client.Tx(ctx)
	require.NoError(t, err)

	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	status, err := repo.GetByName(ctx, statusName)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	require.Equal(t, 1, status.ID)
	require.Equal(t, statusName, status.Name)
}

func TestEquipmentStatusNameRepository_Delete(t *testing.T) {
	ctx := context.Background()
	client := getClient(t, equipmentStatusNameEntityName)
	statusName := "test"
	_, err := client.EquipmentStatusName.Create().SetName(statusName).Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	repo := NewEquipmentStatusNameRepository()
	tx, err := client.Tx(ctx)
	require.NoError(t, err)

	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	status, err := repo.Delete(ctx, 1)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	require.Equal(t, 1, status.ID)
	require.Equal(t, statusName, status.Name)

	selectedStatus, err := client.EquipmentStatusName.Get(ctx, 1)
	require.ErrorContains(t, err, "ent: equipment_status_name not found")
	require.Nil(t, selectedStatus)
}
