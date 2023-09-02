package repo

import (
	"github.com/kozalosev/SadFavBot/db/dto"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/settings"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFetchLanguage(t *testing.T) {
	clearDatabase(t)

	userService := NewUserService(test.BuildApplicationEnv(db))
	lang, _ := userService.FetchUserOptions(TestUID, "en")
	assert.Equal(t, settings.LangCode("en"), lang)

	res, err := db.Exec(ctx, "INSERT INTO users(uid, language) VALUES ($1, 'ru')", TestUID)
	assert.NoError(t, err)
	assert.True(t, res.RowsAffected() > 0)

	lang, _ = userService.FetchUserOptions(TestUID, "en")
	assert.Equal(t, settings.LangCode("ru"), lang)
}

func TestFetchUserOptions(t *testing.T) {
	clearDatabase(t)

	userService := NewUserService(test.BuildApplicationEnv(db))
	_, opts := userService.FetchUserOptions(TestUID, "")
	assert.False(t, opts.(*dto.UserOptions).SubstrSearchEnabled)

	res, err := db.Exec(ctx, "INSERT INTO users(uid, substring_search) VALUES ($1, true)", TestUID)
	assert.NoError(t, err)
	assert.True(t, res.RowsAffected() > 0)

	_, opts = userService.FetchUserOptions(TestUID, "")
	assert.True(t, opts.(*dto.UserOptions).SubstrSearchEnabled)
}

func clearDatabase(t *testing.T) {
	for _, table := range []string{"favs", "links", "packages", "alias_visibility", "users"} {
		//goland:noinspection SqlWithoutWhere
		_, err := db.Exec(ctx, "DELETE FROM "+table)
		assert.NoError(t, err)
	}
}
