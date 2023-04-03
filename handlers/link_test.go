package handlers

import (
	"database/sql"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLinkActionOnConflict(t *testing.T) {
	insertTestData(db)

	reqenv := buildRequestEnv()
	msg := buildMessage(TestUID)
	fields := wizard.Fields{
		&wizard.Field{
			Name: FieldName,
			Data: TestAlias2,
		},
		&wizard.Field{
			Name: FieldAlias,
			Data: TestAlias,
		},
	}

	links, err := fetchLinks(db, TestUID)
	assert.NoError(t, err)
	assert.Len(t, links, 0)

	// no rows will be inserted since TestUID already has an item with TestAlias2
	linkAction(reqenv, msg, fields)

	links, err = fetchLinks(db, TestUID)
	assert.NoError(t, err)
	assert.Len(t, links, 0)
}

func TestLinkAction(t *testing.T) {
	insertTestData(db)

	reqenv := buildRequestEnv()
	msg := buildMessage(TestUID)
	fields := wizard.Fields{
		&wizard.Field{
			Name: FieldName,
			Data: TestAlias2,
		},
		&wizard.Field{
			Name: FieldAlias,
			Data: TestAlias,
		},
	}

	// resolve conflict
	_, err := db.Exec("DELETE FROM items WHERE uid = $1 AND alias = $2", TestUID, TestAlias2ID)
	assert.NoError(t, err)

	links, err := fetchLinks(db, TestUID)
	assert.NoError(t, err)
	assert.Len(t, links, 0)

	// no rows will be inserted since TestUID already has an item with TestAlias2
	linkAction(reqenv, msg, fields)

	links, err = fetchLinks(db, TestUID)
	assert.NoError(t, err)
	assert.Len(t, links, 1)
	assert.Equal(t, TestAlias2+" -> "+TestAlias, links[0])
}

func fetchLinks(db *sql.DB, uid int64) ([]string, error) {
	if rows, err := db.Query("SELECT a1.name || ' -> ' || a2.name FROM links l JOIN aliases a1 ON l.alias_id = a1.id JOIN aliases a2 ON l.linked_alias_id = a2.id WHERE uid = $1", uid); err == nil {
		var (
			links []string
			link  string
		)
		for rows.Next() {
			if err := rows.Scan(&link); err != nil {
				return nil, err
			}
			links = append(links, link)
		}
		return links, nil
	} else {
		return nil, err
	}
}
