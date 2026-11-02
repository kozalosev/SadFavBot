package common

import (
	"testing"
	"time"

	"github.com/kozalosev/goSadTgBot/wizard"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupCleanupPolicy(t *testing.T) {
	policy := GroupCleanupPolicy(5 * time.Minute)
	dialogue := wizard.GroupDialogue{
		Stage:          wizard.StageCompleted,
		ChatID:         -100,
		QuestionIDs:    []int{2, 4},
		UserMessageIDs: []int{1, 3},
		Lang:           testLang,
	}

	assert.Empty(t, policy(dialogue).Delete, "nothing is deleted without the rights")

	dialogue.CanDeleteOthers = true
	cleanup := policy(dialogue)
	assert.ElementsMatch(t, []int{1, 2, 3, 4}, cleanup.Delete, "the whole dialogue")
	assert.Equal(t, selfDeletionDelay, cleanup.After)
	require.NotNil(t, cleanup.Warning)
	assert.Equal(t, 4, cleanup.Warning.MessageID, "only the last question is edited")
	assert.Equal(t, selfDeletionWarning, cleanup.Warning.Grace)

	dialogue.Stage = wizard.StageWaiting
	assert.Equal(t, 5*time.Minute, policy(dialogue).After, "an abandoned dialogue lives as long as the form")
}
