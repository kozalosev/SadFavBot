package handlers

import (
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeleteFormAction(t *testing.T) {
	test.InsertTestData(db)

	msg := buildMessage(test.UID)
	reqenv := test.BuildRequestEnv()
	fields := wizard.Fields{
		&wizard.Field{Name: FieldAlias, Data: test.Alias},
		&wizard.Field{Name: FieldDeleteAll, Data: No},
		&wizard.Field{Name: FieldObject, Type: test.Type, Data: wizard.File{UniqueID: test.UniqueFileID}},
	}

	handler := NewDeleteHandler(test.BuildApplicationEnv(db), nil)
	handler.deleteFormAction(reqenv, msg, fields)

	testAlias := test.Alias
	test.CheckRowsCount(t, db, 1, test.UID, &testAlias) // row with FileID_2 is on its place
	test.CheckRowsCount(t, db, 2, test.UID, nil)        // rows with alias2 and alias+FileID_2

	fields.FindField(FieldDeleteAll).Data = Yes
	handler.deleteFormAction(reqenv, msg, fields)

	test.CheckRowsCount(t, db, 0, test.UID, &testAlias)
}

func TestDeleteFormActionText(t *testing.T) {
	test.InsertTestData(db)

	msg := buildMessage(test.UID2)
	reqenv := test.BuildRequestEnv()
	fields := wizard.Fields{
		&wizard.Field{Name: FieldAlias, Data: test.Alias2},
		&wizard.Field{Name: FieldDeleteAll, Data: No},
		&wizard.Field{Name: FieldObject, Type: wizard.Text, Data: test.Text},
	}

	handler := NewDeleteHandler(test.BuildApplicationEnv(db), nil)
	handler.deleteFormAction(reqenv, msg, fields)

	alias := test.Alias2
	test.CheckRowsCount(t, db, 0, test.UID2, &alias) // row with TestFileID is on its place
}

func TestDeleteFormActionLink(t *testing.T) {
	test.InsertTestData(db)

	_, err := db.Exec(ctx, "DELETE FROM favs WHERE uid = $1 AND alias_id = $2", test.UID2, test.Alias2ID)
	assert.NoError(t, err)
	_, err = db.Exec(ctx, "INSERT INTO links(uid, alias_id, linked_alias_id) VALUES ($1, $2, $3)", test.UID2, test.Alias2ID, test.AliasID)
	assert.NoError(t, err)

	msg := buildMessage(test.UID2)
	reqenv := test.BuildRequestEnv()
	fields := wizard.Fields{
		&wizard.Field{Name: FieldAlias, Data: test.Alias2},
		&wizard.Field{Name: FieldDeleteAll, Data: Yes},
		&wizard.Field{Name: FieldObject},
	}

	assert.Equal(t, 1, checkLinksRowsCount(test.UID2))
	handler := NewDeleteHandler(test.BuildApplicationEnv(db), nil)
	handler.deleteFormAction(reqenv, msg, fields)
	assert.Equal(t, 0, checkLinksRowsCount(test.UID2))
}

func checkLinksRowsCount(uid int64) int {
	var linksRowsCount int
	if err := db.QueryRow(ctx, "SELECT count(id) FROM links WHERE uid = $1", uid).Scan(&linksRowsCount); err == nil {
		return linksRowsCount
	} else {
		return -1
	}
}
