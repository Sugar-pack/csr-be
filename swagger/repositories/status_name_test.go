package repositories

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/enttest"
)

func TestStatusNameRepository_ListOfStatuses(t *testing.T) {
	ctx := context.Background()
	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	statusName := "test"
	_, err := client.StatusName.Create().SetStatus(statusName).Save(ctx)
	if err != nil {
		t.Error(err)
	}
	defer client.Close()

	repo := NewStatusNameRepository(client)
	statuses, err := repo.ListOfStatuses(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, statuses[0].ID)
	assert.Equal(t, statusName, statuses[0].Status)
}
