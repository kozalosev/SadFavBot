package handlers

import (
	"github.com/kozalosev/SadFavBot/db/repo"
	"github.com/kozalosev/SadFavBot/test"
	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRmLinkAction(t *testing.T) {
	test.InsertTestData(db)

	linkName := "link"
	appenv := test.BuildApplicationEnv(db)
	reqenv := test.BuildRequestEnv()
	msg := buildMessage(test.UID)
	fields := wizard.Fields{
		test.NewTextField(FieldName, linkName),
	}

	links, err := fetchLinks(db, test.UID)
	assert.NoError(t, err)
	assert.Len(t, links, 0)

	linkService := repo.NewLinkService(appenv)
	err = linkService.Create(test.UID, linkName, test.Alias)
	assert.NoError(t, err)

	links, err = fetchLinks(db, test.UID)
	assert.NoError(t, err)
	assert.Len(t, links, 1)
	assert.Equal(t, linkName+" -> "+test.Alias, links[0])

	handler := NewRemoveLinkHandler(appenv, nil)
	handler.rmLinkAction(reqenv, msg, fields)

	links, err = fetchLinks(db, test.UID)
	assert.NoError(t, err)
	assert.Len(t, links, 0)
}
