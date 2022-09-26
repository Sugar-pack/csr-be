package repositories

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/enttest"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/swagger/middlewares"
)

func TestStatusNameRepository_ListOfStatuses(t *testing.T) {
	ctx := context.Background()
	client := enttest.Open(t, "sqlite3", "file:statusname?mode=memory&cache=shared&_fk=1")
	statusName := "test"
	_, err := client.StatusName.Create().SetStatus(statusName).Save(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	repo := NewStatusNameRepository()
	tx, err := client.Tx(ctx)
	assert.NoError(t, err)
	ctx = context.WithValue(ctx, middlewares.TxContextKey, tx)
	statuses, err := repo.ListOfStatuses(ctx)
	assert.NoError(t, err)
	assert.NoError(t, tx.Commit())
	assert.Equal(t, 1, statuses[0].ID)
	assert.Equal(t, statusName, statuses[0].Status)
	_, err = client.StatusName.Delete().Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
}
