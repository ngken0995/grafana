package folderimpl

import (
	"context"

	"github.com/grafana/grafana/pkg/infra/db"
	"github.com/grafana/grafana/pkg/services/folder"
)

/*
type HierarchicalEntity[T any] struct {
	a     T
	Left  int64
	Right int64
}

func (h *HierarchicalEntity[T]) GetLeft() int64 {
	return h.Left
}

func (h *HierarchicalEntity[T]) GetRight() int64 {
	return h.Right
}

func (h *HierarchicalEntity[T]) GetEntity() T {
	return h.a
}
*/

type HierarchicalStore struct {
	db db.DB
}

func ProvideHierarchicalStore(db db.DB) *HierarchicalStore {
	store := &HierarchicalStore{db: db}

	//store.populateLeftRightCols(1, nil, 0, 0)
	return store
}

func (hs *HierarchicalStore) migrate(ctx context.Context, orgID int64, f *folder.Folder, counter int64) (int64, error) {
	// TODO: run only once
	err := hs.db.InTransaction(ctx, func(ctx context.Context) error {
		var children []*folder.Folder

		q := "SELECT * FROM folder WHERE org_id = ?"
		args := []interface{}{orgID}
		// get children
		if f == nil {
			q = q + "AND parent_uid IS NULL"
		} else {
			q = q + "AND parent_uid = ?"
			args = append(args, f.UID)
		}

		if err := hs.db.WithDbSession(ctx, func(sess *db.Session) error {
			if err := sess.SQL(q, args...).Find(&children); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		if len(children) == 0 {
			counter++
			if f != nil {
				f.Rgt = counter
			}
			return nil
		}

		for _, child := range children {
			counter++
			child.Lft = counter
			c, err := hs.migrate(ctx, orgID, child, counter)
			if err != nil {
				return err
			}
			if err := hs.db.WithDbSession(ctx, func(sess *db.Session) error {
				_, err := sess.Exec("UPDATE folder SET lft = ?, rgt = ? WHERE uid = ? AND org_id = ?", child.Lft, child.Rgt, child.UID, child.OrgID)
				if err != nil {
					return err
				}
				return nil
			}); err != nil {
				return err
			}
			counter = c
		}
		counter++
		if f != nil {
			f.Rgt = counter
		}
		return nil
	})

	return counter, err
}

func (hs *HierarchicalStore) Create(ctx context.Context, cmd folder.CreateFolderCommand) (*folder.Folder, error) {
	panic("not implemented")
}

func (hs *HierarchicalStore) Delete(ctx context.Context, uid string, orgID int64) error {
	panic("not implemented")
}

func (hs *HierarchicalStore) Update(ctx context.Context, cmd folder.UpdateFolderCommand) (*folder.Folder, error) {
	panic("not implemented")
}

func (hs *HierarchicalStore) Get(ctx context.Context, cmd folder.GetFolderQuery) (*folder.Folder, error) {
	panic("not implemented")
}

func (hs *HierarchicalStore) GetParents(ctx context.Context, cmd folder.GetParentsQuery) ([]*folder.Folder, error) {
	var folders []*folder.Folder
	err := hs.db.WithDbSession(ctx, func(sess *db.Session) error {
		if err := sess.SQL(`SELECT parent.*
		FROM folder AS node,
			folder AS parent
		WHERE node.lft > parent.lft AND node.lft < parent.rgt
			AND node.org_id = ? AND node.uid = ?
		ORDER BY node.lft`, cmd.OrgID, cmd.UID).Find(&folders); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return folders, nil
}

func (hs *HierarchicalStore) GetChildren(ctx context.Context, cmd folder.GetChildrenQuery) ([]*folder.Folder, error) {
	panic("not implemented")
}

func (hs *HierarchicalStore) GetHeight(ctx context.Context, foldrUID string, orgID int64, parentUID *string) (int, error) {
	panic("not implemented")
}
