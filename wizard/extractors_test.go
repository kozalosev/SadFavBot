package wizard

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"

	"reflect"
	"testing"
)

func TestRestoreExtractor(t *testing.T) {
	audioMsg := &tgbotapi.Message{
		Audio: &tgbotapi.Audio{FileID: TestFileID, FileUniqueID: TestFileUniqueID},
	}

	f := Field{Type: Image}
	f.restoreExtractor(audioMsg)
	assert.Equal(t, getFuncPtr(imageExtractor), getFuncPtr(f.extractor))

	f = Field{Type: Auto}
	f.restoreExtractor(audioMsg)
	assert.Equal(t, getFuncPtr(audioExtractor), getFuncPtr(f.extractor))
	expectedAudio := File{ID: TestFileID, UniqueID: TestFileUniqueID}
	assert.Equal(t, expectedAudio, f.extractor(audioMsg))
}

func getFuncPtr(f interface{}) uintptr {
	return reflect.ValueOf(f).Pointer()
}
