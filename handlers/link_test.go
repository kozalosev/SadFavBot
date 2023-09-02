package handlers

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLinkActionOnConflict(t *testing.T) {
	test.InsertTestData(db)

	appenv := test.BuildApplicationEnv(db)
	reqenv := test.BuildRequestEnv()
	msg := buildMessage(test.UID)
	fields := wizard.Fields{
		test.NewTextField(FieldName, test.Alias2),
		test.NewTextField(FieldAlias, test.Alias),
	}

	links, err := fetchLinks(db, test.UID)
	assert.NoError(t, err)
	assert.Len(t, links, 0)

	handler := NewLinkHandler(appenv, nil)
	// no rows will be inserted since TestUID already has an item with TestAlias2
	handler.linkAction(reqenv, msg, fields)

	links, err = fetchLinks(db, test.UID)
	assert.NoError(t, err)
	assert.Len(t, links, 0)
}

func TestLinkAction(t *testing.T) {
	test.InsertTestData(db)

	appenv := test.BuildApplicationEnv(db)
	reqenv := test.BuildRequestEnv()
	msg := buildMessage(test.UID)
	fields := wizard.Fields{
		test.NewTextField(FieldName, test.Alias2),
		test.NewTextField(FieldAlias, test.Alias),
	}

	// resolve conflict
	_, err := db.Exec(ctx, "DELETE FROM favs WHERE uid = $1 AND alias_id = $2", test.UID, test.Alias2ID)
	assert.NoError(t, err)

	links, err := fetchLinks(db, test.UID)
	assert.NoError(t, err)
	assert.Len(t, links, 0)

	handler := NewLinkHandler(appenv, nil)
	handler.linkAction(reqenv, msg, fields)

	links, err = fetchLinks(db, test.UID)
	assert.NoError(t, err)
	assert.Len(t, links, 1)
	assert.Equal(t, test.Alias2+" -> "+test.Alias, links[0])
}

func fetchLinks(db *pgxpool.Pool, uid int64) ([]string, error) {
	if rows, err := db.Query(ctx, "SELECT a1.name || ' -> ' || a2.name FROM links l JOIN aliases a1 ON l.alias_id = a1.id JOIN aliases a2 ON l.linked_alias_id = a2.id WHERE uid = $1", uid); err == nil {
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
