package handlers

import (
	"github.com/kozalosev/SadFavBot/db/dto"
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
