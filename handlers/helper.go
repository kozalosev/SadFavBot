package handlers

import (
	"database/sql"
	"github.com/kozalosev/SadFavBot/base"
	"github.com/kozalosev/SadFavBot/wizard"
	log "github.com/sirupsen/logrus"
)

type itemValues struct {
	Alias string
	Type  wizard.FieldType
	Text  string
	File  *wizard.File
}

func extractItemValues(fields wizard.Fields) (*itemValues, bool) {
	aliasField := fields.FindField(FieldAlias)
	objectField := fields.FindField(FieldObject)

	alias, ok := aliasField.Data.(string)
	if !ok {
		log.Errorf("Invalid type for alias: %T %+v", aliasField, aliasField)
		return nil, false
	}

	if objectField.Data == nil {
		return &itemValues{Alias: alias}, true
	}

	var (
		text string
		file wizard.File
	)
	if objectField.Type == wizard.Text {
		text, ok = objectField.Data.(string)
		if !ok {
			log.Errorf("Invalid type: string was expected but '%T %+v' is got", objectField.Data, objectField.Data)
			return nil, false
		}
	} else {
		file, ok = objectField.Data.(wizard.File)
		if !ok {
			log.Errorf("Invalid type: File was expected but '%T %+v' is got", objectField.Data, objectField.Data)
			return nil, false
		}
	}

	return &itemValues{
		Alias: alias,
		Type:  objectField.Type,
		Text:  text,
		File:  &file,
	}, true
}

func checkRowsWereAffected(res sql.Result) bool {
	var (
		rowsAffected int64
		err          error
	)
	if rowsAffected, err = res.RowsAffected(); err != nil {
		log.Errorln(err)
		rowsAffected = -1 // logs but ignores
	}
	if rowsAffected == 0 {
		return false
	} else {
		return true
	}
}

func replierFactory(reqenv *base.RequestEnv) func(string) {
	return func(statusKey string) {
		reqenv.Reply(reqenv.Lang.Tr(statusKey))
	}
}
