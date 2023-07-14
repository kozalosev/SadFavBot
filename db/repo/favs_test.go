package repo

import (
	"github.com/kozalosev/SadFavBot/db/dto"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFavService_Find(t *testing.T) {
	test.InsertTestData(db)

	query := test.BuildInlineQuery()
	favsService := NewFavsService(test.BuildApplicationEnv(db))
	objects, err := favsService.Find(query.From.ID, query.Query, false)

	assert.NoError(t, err)
	assert.Len(t, objects, 2)
	assert.Equal(t, test.FileID, objects[0].File.ID)
	assert.Equal(t, test.FileID2, objects[1].File.ID)
}

func TestFavService_Find_Text(t *testing.T) {
	test.InsertTestData(db)

	favsService := NewFavsService(test.BuildApplicationEnv(db))
	objects, err := favsService.Find(test.UID2, test.Alias2, false)

	assert.NoError(t, err)
	assert.Len(t, objects, 1)
	assert.Equal(t, test.Text, *objects[0].Text)
}

func TestFavService_Find_Location(t *testing.T) {
	test.InsertTestData(db)
	aliasLoc := test.Alias + "Loc"
	favsService := NewFavsService(test.BuildApplicationEnv(db))
	res, err := favsService.Save(test.UID2, aliasLoc, &dto.Fav{
		Type:     wizard.Location,
		Location: &wizard.LocData{Latitude: test.Latitude, Longitude: test.Longitude},
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), res.RowsAffected())

	objects, err := favsService.Find(test.UID2, aliasLoc, false)

	assert.NoError(t, err)
	assert.Len(t, objects, 1)
	assert.Equal(t, test.Latitude, objects[0].Location.Latitude)
	assert.Equal(t, test.Longitude, objects[0].Location.Longitude)
}

func TestFavService_Find_bySubstring(t *testing.T) {
	test.InsertTestData(db)

	query := test.BuildInlineQuery()
	query.Query = "a"

	favsService := NewFavsService(test.BuildApplicationEnv(db))
	objects, err := favsService.Find(query.From.ID, query.Query, false)

	assert.NoError(t, err)
	assert.Len(t, objects, 0)

	objects, err = favsService.Find(query.From.ID, query.Query, true)

	assert.Len(t, objects, 2)
	assert.Equal(t, test.FileID, objects[0].File.ID)
	assert.Equal(t, test.FileID2, objects[1].File.ID)
}

func TestFavService_Find_escaping(t *testing.T) {
	test.InsertTestData(db)

	query := test.BuildInlineQuery()
	query.Query = "%a%"

	favsService := NewFavsService(test.BuildApplicationEnv(db))
	objects, err := favsService.Find(query.From.ID, query.Query, false)

	assert.NoError(t, err)
	assert.Len(t, objects, 0)
}

func TestFavService_Find_byLink(t *testing.T) {
	test.InsertTestData(db)

	_, err := db.Exec(ctx, "DELETE FROM favs WHERE uid = $1 AND alias_id = $2", test.UID, test.AliasID)
	assert.NoError(t, err)
	_, err = db.Exec(ctx, "INSERT INTO links(uid, alias_id, linked_alias_id) VALUES ($1, $2, $3)", test.UID, test.AliasID, test.Alias2ID)
	assert.NoError(t, err)

	query := test.BuildInlineQuery()

	favsService := NewFavsService(test.BuildApplicationEnv(db))
	objects, err := favsService.Find(query.From.ID, query.Query, false)

	assert.Len(t, objects, 1)
	assert.Equal(t, test.FileID, objects[0].File.ID)
}
