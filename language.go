package main

import (
	"database/sql"
	log "github.com/sirupsen/logrus"
)

func fetchLanguage(db *sql.DB, uid int64, defaultLang string) string {
	res := db.QueryRow("SELECT language FROM users WHERE uid = $1", uid)
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
