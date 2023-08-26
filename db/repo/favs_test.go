package repo

import (
	"github.com/kozalosev/SadFavBot/db/dto"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
	"github.com/thoas/go-funk"
	"testing"
)

func TestFavService_Find(t *testing.T) {
	test.InsertTestData(db)
	testFindImpl(t)
}

func TestFavService_Find_hidden(t *testing.T) {
	test.InsertTestData(db)

	appEnv := test.BuildApplicationEnv(db)
	aliasService := NewAliasService(appEnv)
	err := aliasService.Hide(test.UID, test.Alias)
	assert.NoError(t, err)

	testFindImpl(t)
}

func testFindImpl(t *testing.T) {
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

func TestFavService_Find_bySubstring_hidden(t *testing.T) {
	test.InsertTestData(db)

	appEnv := test.BuildApplicationEnv(db)
	aliasService := NewAliasService(appEnv)
	err := aliasService.Hide(test.UID, test.Alias)
	assert.NoError(t, err)

	query := test.BuildInlineQuery()
	query.Query = "a"

	favsService := NewFavsService(test.BuildApplicationEnv(db))
	objects, err := favsService.Find(query.From.ID, query.Query, true)

	assert.Len(t, objects, 1)
	assert.Equal(t, test.FileID, objects[0].File.ID)
}

func TestFavService_Find_bySubstring_withDuplicates(t *testing.T) {
	test.InsertTestData(db)

	query := test.BuildInlineQuery()
	query.From.ID = test.UID2
	query.Query = "a"

	dupFile := &wizard.File{
		ID:       test.FileID2,
		UniqueID: test.UniqueFileID,
	}
	dupFavFile := &dto.Fav{
		Type: wizard.Sticker,
		File: dupFile,
	}
	dupText := test.Text + "'2"
	dupFavText := &dto.Fav{
		Type: wizard.Text,
		Text: &dupText,
	}
	favsService := NewFavsService(test.BuildApplicationEnv(db))
	for _, dupFav := range []*dto.Fav{dupFavFile, dupFavText} {
		rowsAffectedAware, err := favsService.Save(query.From.ID, query.Query, dupFav)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), rowsAffectedAware.RowsAffected())
	}

	objects, err := favsService.Find(query.From.ID, query.Query, true)
	assert.NoError(t, err)
	assert.Len(t, objects, 3)

	types := funk.Map(objects, func(fav *dto.Fav) wizard.FieldType {
		return fav.Type
	}).([]wizard.FieldType)
	assert.Contains(t, types, wizard.Sticker)
	assert.Contains(t, types, wizard.Text)
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

	appenv := test.BuildApplicationEnv(db)
	favsService := NewFavsService(appenv)
	objects, err := favsService.Find(query.From.ID, query.Query, false)

	assert.NoError(t, err)
	assert.Len(t, objects, 1)
	assert.Equal(t, test.FileID, objects[0].File.ID)

	// several links beginning alike + substring search
	// https://github.com/kozalosev/SadFavBot/issues/75

	linkService := NewLinkService(appenv)
	oleas := "oleas"
	for _, link := range []string{oleas, oleas + "'2"} {
		err = linkService.Create(test.UID, link, test.Alias2)
		assert.NoError(t, err)
	}

	objects, err = favsService.Find(query.From.ID, oleas, true)

	assert.NoError(t, err)
	assert.Len(t, objects, 1)
	assert.Equal(t, test.FileID, objects[0].File.ID)
}

func TestFavService_Find_PhotoWithCaption(t *testing.T) {
	test.InsertTestData(db)

	query := test.BuildInlineQuery()
	query.From.ID = test.UID3
	fav := &dto.Fav{
		Type: wizard.Image,
		File: &wizard.File{
			ID:       test.FileIDPhoto,
			UniqueID: test.FileIDPhoto,
			Caption:  test.AliasPhoto,
		},
	}

	appenv := test.BuildApplicationEnv(db)
	favsService := NewFavsService(appenv)
	rowsAffectedAware, err := favsService.Save(query.From.ID, test.AliasPhoto, fav)
	assert.NoError(t, err)
	assert.Greater(t, rowsAffectedAware.RowsAffected(), int64(0))

	objects, err := favsService.Find(query.From.ID, test.AliasPhoto, false)

	assert.NoError(t, err)
	assert.Len(t, objects, 1)
	assert.Equal(t, test.FileIDPhoto, objects[0].File.ID)
	assert.Equal(t, test.CaptionPhoto, objects[0].File.Caption)
}
