package handlers

import (
	"github.com/kozalosev/SadFavBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	en = "en"
	ru = "ru"
)

func TestLanguageFormAction(t *testing.T) {
	insertTestData(db)
	assertLanguage(t, "ru")

	msg := buildMessage(TestUID)
	reqenv := buildRequestEnv()
	fields := wizard.Fields{
		&wizard.Field{
			Name: FieldLanguage,
			Data: en,
		},
	}

	languageFormAction(reqenv, msg, fields)
	assertLanguage(t, en)

	fields.FindField(FieldLanguage).Data = "au" // unsupported
	languageFormAction(reqenv, msg, fields)
	assertLanguage(t, en)
}

func TestFlagToCode(t *testing.T) {
	assert.Equal(t, en, langFlagToCode("🇺🇸"))
	assert.Equal(t, en, langFlagToCode(en))
}

func assertLanguage(t *testing.T, expected string) {
	res := db.QueryRow("SELECT language FROM users WHERE uid = $1", TestUID)
	var lang string
	err := res.Scan(&lang)

	assert.NoError(t, err)
	assert.Equal(t, expected, lang)
}
