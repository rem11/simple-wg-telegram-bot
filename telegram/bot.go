package telegram

import (
	"log"
	"time"

	"github.com/rem11/simple-wg-telegram-bot/wireguard"
	"gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

type Bot struct {
	ConfigManager  *wireguard.ConfigManager
	PollingTimeout time.Duration
	*CommandController
	Token   string
	UserIDs []int64
}

func handleError(err error, ctx telebot.Context) {
	log.Println(err)
}

func (bot *Bot) Start() error {
	pref := telebot.Settings{
		Token: bot.Token,
		Poller: &telebot.LongPoller{
			Timeout: bot.PollingTimeout,
		},
		OnError: handleError,
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		return err
	}

	b.Use(middleware.Whitelist(bot.UserIDs...))

	b.Handle("/add_peer", func(ctx telebot.Context) error {
		bot.CommandController.Start(&AddPeerCommand{
			ConfigManager: bot.ConfigManager,
		}, ctx)
		return nil
	})

	b.Handle("/remove_peer", func(ctx telebot.Context) error {
		bot.CommandController.Start(&RemovePeerCommand{
			ConfigManager: bot.ConfigManager,
		}, ctx)
		return nil
	})

	b.Handle("/client_config", func(ctx telebot.Context) error {
		bot.CommandController.Start(&ClientConfigCommand{
			ConfigManager: bot.ConfigManager,
		}, ctx)
		return nil
	})

	b.Handle(telebot.OnText, func(ctx telebot.Context) error {
		bot.CommandController.HandleInput(ctx)
		return nil
	})

	b.SetCommands([]telebot.Command{
		{
			Text:        "add_peer",
			Description: "Add new peer to server configuration",
		},
		{
			Text:        "remove_peer",
			Description: "Remove peer from server configuration",
		},
		{
			Text:        "client_config",
			Description: "Get client config for the specific peer",
		},
	})

	b.Start()

	return nil
}
