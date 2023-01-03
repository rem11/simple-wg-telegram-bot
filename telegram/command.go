package telegram

import (
	"gopkg.in/telebot.v3"
)

type Command interface {
	Start(telebot.Context) bool
	HandleInput(telebot.Context) bool
}
