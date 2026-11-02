package handlers

import (
	"testing"

	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/stretchr/testify/assert"
)

func TestIgnoreForeignCommandsHandler_CanHandle(t *testing.T) {
	handler := NewIgnoreForeignCommandsHandler("SadFavBot")

	tests := []struct {
		text     string
		expected bool
	}{
		{"/help", false},
		{"/help@SadFavBot", false},
		{"/help@sadfavbot", false},
		{"/help@AnotherBot", true},
		{"/help@AnotherBot arg", true},
		{"some text @AnotherBot", false},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			msg := buildMessage(1)
			msg.Text = tt.text
			if tt.text[0] == '/' {
				cmdLen := len(tt.text)
				for i, c := range tt.text {
					if c == ' ' {
						cmdLen = i
						break
					}
				}
				msg.Entities = []tgbotapi.MessageEntity{{Type: "bot_command", Offset: 0, Length: cmdLen}}
			}
			assert.Equal(t, tt.expected, handler.CanHandle(nil, msg))
		})
	}
}
