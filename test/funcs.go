package test

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kozalosev/SadFavBot/db/dto"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/loctools/go-l10n/loc"
	"github.com/stretchr/testify/assert"
	"testing"
)

var ctx = context.Background()

func InsertTestData(db *pgxpool.Pool) {
	for _, table := range []string{"alias_visibility", "links", "package_aliases", "packages", "favs", "aliases", "texts", "users"} {
		_, err := db.Exec(ctx, "DELETE FROM "+table)
		check(err)
	}

	_, err := db.Exec(ctx, "INSERT INTO aliases(id, name) VALUES ($1, $2), ($3, $4)",
		AliasID, Alias, Alias2ID, Alias2)
	check(err)
	_, err = db.Exec(ctx, "ALTER SEQUENCE aliases_id_seq RESTART WITH 3")
	check(err)
	_, err = db.Exec(ctx, "INSERT INTO favs(uid, type, alias_id, file_id, file_unique_id) VALUES"+
		"($1, $3, $4, $6, $8),"+ // TestUID, TestAlias, TestFileID, TestUniqueFileID
		"($1, $3, $4, $7, $9),"+ // TestUID, TestAlias, TestFileID2, TestUniqueFileID2
		"($1, $3, $5, $6, $8),"+ // TestUID, TestAlias2, TestFileID, TestUniqueFileID
		"($2, $3, $4, $6, $8)", // TestUID2, TestAlias, TestFileID, TestUniqueFileID
		UID, UID2, Type, AliasID, Alias2ID, FileID, FileID2, UniqueFileID, UniqueFileID2)
	check(err)
	_, err = db.Exec(ctx, "INSERT INTO texts(id, text) VALUES ($1, $2)", TextID, Text)
	check(err)
	_, err = db.Exec(ctx, "INSERT INTO favs(uid, type, alias_id, text_id) VALUES ($1, $2, $3, $4)",
		UID2, wizard.Text, Alias2ID, TextID)
	check(err)

	_, err = db.Exec(ctx, "INSERT INTO users(uid, language) VALUES ($1, 'ru'), ($2, 'en'), ($3, 'ru')", UID, UID2, UID3)
}

func InsertTestPackages(db *pgxpool.Pool) {
	_, err := db.Exec(ctx, "INSERT INTO packages(id, owner_uid, name) VALUES ($1, $2, $3)", PackageID, UID, Package)
	check(err)
	_, err = db.Exec(ctx, "INSERT INTO package_aliases(package_id, alias_id) VALUES ($1, $2)", PackageID, Alias2ID)
	check(err)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func CheckRowsCount(t *testing.T, db *pgxpool.Pool, expected int, uid int64, alias *string) {
	var countRes pgx.Row
	if alias != nil {
		countRes = db.QueryRow(ctx, "SELECT count(id) FROM favs WHERE uid = $1 AND alias_id = (SELECT id FROM aliases WHERE name = $2)", uid, alias)
	} else {
		countRes = db.QueryRow(ctx, "SELECT count(id) FROM favs WHERE uid = $1", uid)
	}
	var count int
	err := countRes.Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, expected, count)
}

func BuildApplicationEnv(db *pgxpool.Pool) *base.ApplicationEnv {
	return &base.ApplicationEnv{
		Bot:      &base.FakeBotAPI{},
		Database: db,
		Ctx:      ctx,
	}
}

func BuildRequestEnv() *base.RequestEnv {
	return &base.RequestEnv{
		Lang:    loc.NewPool("en").GetContext("en"),
		Options: &dto.UserOptions{},
	}
}

func BuildInlineQuery() *tgbotapi.InlineQuery {
	return &tgbotapi.InlineQuery{
		From:  &tgbotapi.User{ID: UID},
		Query: AliasCI,
	}
}
