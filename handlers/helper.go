package handlers

import (
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kozalosev/SadFavBot/db/dto"
	"github.com/kozalosev/goSadTgBot/base"
	"github.com/kozalosev/goSadTgBot/logconst"
	"github.com/kozalosev/goSadTgBot/wizard"
	log "github.com/sirupsen/logrus"
	"strings"
)

var markdownEscaper = strings.NewReplacer(
	"*", "\\*",
	"_", "\\_",
	"`", "\\`")

func extractFavInfo(fields wizard.Fields) (string, *dto.Fav) {
	aliasField := fields.FindField(FieldAlias)
	objectField := fields.FindField(FieldObject)

	alias, ok := aliasField.Data.(string)
	if !ok {
		log.WithField(logconst.FieldFunc, "extractFavInfo").
			Errorf("Invalid type for alias: %T %+v", aliasField, aliasField)
		return "", nil
	}

	if objectField.Data == nil {
		return alias, &dto.Fav{}
	}

	var (
		text string
		file wizard.File
	)
	if objectField.Type == wizard.Text {
		text, ok = objectField.Data.(string)
		if !ok {
			log.WithField(logconst.FieldFunc, "extractFavInfo").
				Errorf("Invalid type: string was expected but '%T %+v' is got", objectField.Data, objectField.Data)
			return "", nil
		}
	} else {
		file, ok = objectField.Data.(wizard.File)
		if !ok {
			log.WithField(logconst.FieldFunc, "extractFavInfo").
				Errorf("Invalid type: File was expected but '%T %+v' is got", objectField.Data, objectField.Data)
			return "", nil
		}
	}

	return alias, &dto.Fav{
		Type: objectField.Type,
		Text: &text,
		File: &file,
	}
}

func isDuplicateConstraintViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == DuplicateConstraintSQLCode
}

func replierFactory(appenv *base.ApplicationEnv, reqenv *base.RequestEnv, msg *tgbotapi.Message) func(string) {
	return func(statusKey string) {
		appenv.Bot.Reply(msg, reqenv.Lang.Tr(statusKey))
	}
}
