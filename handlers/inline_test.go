package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/loctools/go-l10n/loc"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestFindObjects(t *testing.T) {
	insertTestData(db)

	query := buildInlineQuery()
	reqenv := buildRequestEnv()
	objects := findObjects(reqenv, query)

	assert.Len(t, objects, 2)
	assert.Equal(t, TestFileID, *objects[0].FileID)
	assert.Equal(t, TestFileID2, *objects[1].FileID)
}

func TestFindObjectsBySubstring(t *testing.T) {
	insertTestData(db)

	query := buildInlineQuery()
	query.Query = "a"
	reqenv := buildRequestEnv()

	objects := findObjects(reqenv, query)

	assert.Len(t, objects, 0)

	reqenv.Options.SubstrSearchEnabled = true
	objects = findObjects(reqenv, query)

	assert.Len(t, objects, 2)
	assert.Equal(t, TestFileID, *objects[0].FileID)
	assert.Equal(t, TestFileID2, *objects[1].FileID)
}

func TestFindObjectsEscaping(t *testing.T) {
	insertTestData(db)

	query := buildInlineQuery()
	query.Query = "%a%"
	reqenv := buildRequestEnv()

	objects := findObjects(reqenv, query)

	assert.Len(t, objects, 0)
}

func TestFindObjectsByLink(t *testing.T) {
	insertTestData(db)

	_, err := db.Exec("DELETE FROM items WHERE uid = $1 AND alias = $2", TestUID, TestAliasID)
	assert.NoError(t, err)
	_, err = db.Exec("INSERT INTO links(uid, alias_id, linked_alias_id) VALUES ($1, $2, $3)", TestUID, TestAliasID, TestAlias2ID)
	assert.NoError(t, err)

	query := buildInlineQuery()
	reqenv := buildRequestEnv()
	objects := findObjects(reqenv, query)

	assert.Len(t, objects, 1)
	assert.Equal(t, TestFileID, *objects[0].FileID)
}

func TestMapper(t *testing.T) {
	insertTestData(db)

	query := buildInlineQuery()
	reqenv := buildRequestEnv()
	objects := findObjects(reqenv, query)

	inlineAnswer := generateMapper(loc.NewPool("en").GetContext("en"))(objects[0])
	assert.Equal(t, "InlineQueryResultCachedSticker", reflect.TypeOf(inlineAnswer).Name())
}

func buildInlineQuery() *tgbotapi.InlineQuery {
	return &tgbotapi.InlineQuery{
		From:  &tgbotapi.User{ID: TestUID},
		Query: TestAliasCI,
	}
}
