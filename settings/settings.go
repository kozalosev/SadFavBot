package settings

import (
	"context"
	"database/sql"
	log "github.com/sirupsen/logrus"
)

// LangCode is a language code like 'ru' or 'en'.
type LangCode string

// UserOptions stored in the database.
type UserOptions struct {
	SubstrSearchEnabled bool
}

// FetchUserOptions from the database if they exist.
func FetchUserOptions(ctx context.Context, db *sql.DB, uid int64, defaultLang string) (LangCode, *UserOptions) {
	var (
		lang *string
		opts UserOptions
	)
	if err := db.QueryRowContext(ctx, "SELECT language, substring_search FROM users WHERE uid = $1", uid).Scan(&lang, &opts.SubstrSearchEnabled); err != nil {
		log.Error(err)
		return LangCode(defaultLang), &opts
	} else if lang != nil && len(*lang) > 0 {
		return LangCode(*lang), &opts
	} else {
		return LangCode(defaultLang), &opts
	}
}
