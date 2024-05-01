package handlers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func buildMessage(uid int64) *tgbotapi.Message {
	return &tgbotapi.Message{
		From: &tgbotapi.User{ID: uid},
		Chat: tgbotapi.Chat{Type: "private"},
	}
}
