package folderimpl

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/grafana/grafana/pkg/infra/db"
	"github.com/grafana/grafana/pkg/services/folder"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var folders = [][]any{
	// org_id, uid, title, created, updated, parent_uid, left, right
	{1, "1", "ELECTRONICS", time.Now(), time.Now(), nil, 1, 20},
	{1, "2", "TELEVISIONS", time.Now(), time.Now(), "1", 2, 9},
	{1, "3", "TUBE", time.Now(), time.Now(), "2", 3, 4},
	{1, "4", "LCD", time.Now(), time.Now(), "2", 5, 6},
	{1, "5", "PLASMA", time.Now(), time.Now(), "2", 7, 8},
	{1, "6", "PORTABLE ELECTRONICS", time.Now(), time.Now(), "1", 10, 19},
	{1, "7", "MP3 PLAYERS", time.Now(), time.Now(), "6", 11, 14},
	{1, "8", "FLASH", time.Now(), time.Now(), "7", 12, 13},
	{1, "9", "CD PLAYERS", time.Now(), time.Now(), "6", 15, 16},
	{1, "10", "2 WAY RADIOS", time.Now(), time.Now(), "6", 17, 18},
}

// storeFolders stores the folders in the database.
// if storeLeftRight is true, the left and right values are stored as well.
func storeFolders(t *testing.T, storeDB db.DB, storeLeftRight bool) {
	t.Helper()

	storeDB.WithDbSession(context.Background(), func(sess *db.Session) error {
		cols := []string{"org_id", "uid", "title", "created", "updated", "parent_uid"}
		if storeLeftRight {
			cols = append(cols, "lft", "rgt")
		}
		sql := "INSERT INTO folder(" + strings.Join(cols, ",") + ") VALUES"
		sqlOrArgs := []any{}
		for i := 0; i < len(folders); i++ {
			sql = sql + "(" + strings.TrimSuffix(strings.Repeat("?,", len(cols)), ",") + ")"
			if i < len(folders)-1 {
				sql = sql + ","
			}
			sqlOrArgs = append(sqlOrArgs, folders[i][:len(cols)]...)
		}
		sqlOrArgs = append([]any{sql}, sqlOrArgs...)
		spew.Dump(">>>>", sqlOrArgs)

		_, err := sess.Exec(sqlOrArgs...)
		require.NoError(t, err)
		return nil
	})

}

func TestIntegrationMigrate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	sqlStore := sqlstore.InitTestDB(t)

	folderStore := ProvideHierarchicalStore(sqlStore)
	storeFolders(t, folderStore.db, false)

	_, err := folderStore.migrate(context.Background(), 1, nil, 0)
	require.NoError(t, err)

	var r []*folder.Folder
	folderStore.db.WithDbSession(context.Background(), func(sess *db.Session) error {
		return sess.SQL("SELECT * FROM folder").Find(&r)
	})
	require.NoError(t, err)

	for i := 0; i < len(folders); i++ {
		assert.Equal(t, folders[i][0], int(r[i].OrgID))
		assert.Equal(t, folders[i][1], r[i].UID)
		assert.Equal(t, folders[i][6], int(r[i].Lft))
		assert.Equal(t, folders[i][7], int(r[i].Rgt))
	}
}

func TestIntegrationGetParentsMPTT(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	sqlStore := sqlstore.InitTestDB(t)
	folderStore := ProvideHierarchicalStore(sqlStore)
	storeFolders(t, folderStore.db, true)

	ancestors, err := folderStore.GetParents(context.Background(), folder.GetParentsQuery{
		OrgID: 1,
		UID:   "8",
	})
	require.NoError(t, err)

	expected := []string{"ELECTRONICS", "PORTABLE ELECTRONICS", "MP3 PLAYERS"}
	assert.Equal(t, len(expected), len(ancestors))
	for i := 0; i < len(ancestors); i++ {
		assert.Equal(t, expected[i], ancestors[i].Title)
	}
}
