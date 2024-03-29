package handlers

import (
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

const en = "en"

func TestLanguageFormAction(t *testing.T) {
	test.InsertTestData(db)
	assertLanguage(t, "ru")

	msg := buildMessage(test.UID)
	appenv := test.BuildApplicationEnv(db)
	reqenv := test.BuildRequestEnv()
	fields := wizard.Fields{
		test.NewTextField(FieldLanguage, en),
	}

	handler := NewLanguageHandler(appenv, nil)
	handler.languageFormAction(reqenv, msg, fields)
	assertLanguage(t, en)

	fields.FindField(FieldLanguage).Data = wizard.Txt{Value: "au"} // unsupported
	handler.languageFormAction(reqenv, msg, fields)
	assertLanguage(t, en)
}

func TestFlagToCode(t *testing.T) {
	assert.Equal(t, en, langFlagToCode("🇺🇸"))
	assert.Equal(t, en, langFlagToCode(en))
}

func assertLanguage(t *testing.T, expected string) {
	res := db.QueryRow(ctx, "SELECT language FROM users WHERE uid = $1", test.UID)
	var lang string
	err := res.Scan(&lang)

	assert.NoError(t, err)
	assert.Equal(t, expected, lang)
}
