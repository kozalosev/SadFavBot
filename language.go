package main

import (
	"context"
	"database/sql"
	log "github.com/sirupsen/logrus"
)

func fetchLanguage(ctx context.Context, db *sql.DB, uid int64, defaultLang string) string {
	res := db.QueryRowContext(ctx, "SELECT language FROM users WHERE uid = $1", uid)
	if res.Err() != nil {
		log.Errorln(res.Err())
		return defaultLang
	}
	var lang string
	err := res.Scan(&lang)
	if err != nil {
		log.Errorln(err)
		return defaultLang
	}
	if len(lang) > 0 {
		return lang
	} else {
		return defaultLang
	}
}
