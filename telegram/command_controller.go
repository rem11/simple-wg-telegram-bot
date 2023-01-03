package telegram

import (
	"gopkg.in/telebot.v3"
)

type CommandController struct {
	Commands map[int64]Command
}

func NewCommandController() *CommandController {
	return &CommandController{
		Commands: map[int64]Command{},
	}
}

func (cc *CommandController) Start(cmd Command, ctx telebot.Context) {
	result := cmd.Start(ctx)
	if !result {
		cc.Commands[ctx.Chat().ID] = cmd
	}
}

func (cc *CommandController) HandleInput(ctx telebot.Context) {
	cmd := cc.Commands[ctx.Chat().ID]
	if cmd != nil {
		result := cmd.HandleInput(ctx)
		if result {
			delete(cc.Commands, ctx.Chat().ID)
		}
	}
}
