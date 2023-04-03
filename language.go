package main

import (
	"context"
	"database/sql"
	log "github.com/sirupsen/logrus"
)

func fetchLanguage(ctx context.Context, db *sql.DB, uid int64, defaultLang string) string {
	var lang *string
	if err := db.QueryRowContext(ctx, "SELECT language FROM users WHERE uid = $1", uid).Scan(&lang); err != nil {
		log.Error(err)
		return defaultLang
	} else if lang != nil && len(*lang) > 0 {
		return *lang
	} else {
		return defaultLang
	}
}
