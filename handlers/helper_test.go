package handlers

import tgbotapi "github.com/OvyFlash/telegram-bot-api"

func buildMessage(uid int64) *tgbotapi.Message {
	return &tgbotapi.Message{
		From: &tgbotapi.User{ID: uid},
		Chat: tgbotapi.Chat{Type: "private"},
	}
}
