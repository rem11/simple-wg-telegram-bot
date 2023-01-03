package telegram

import (
	"fmt"
	"log"
	"strings"

	"github.com/rem11/simple-wg-telegram-bot/wireguard"
	"gopkg.in/telebot.v3"
)

const addConfirmation = `Are you sure that you want to add new peer?
Public key: %s
Name: %s`

type AddPeerCommand struct {
	*wireguard.ConfigManager
	publicKey string
	name      string
}

func (cmd *AddPeerCommand) Start(ctx telebot.Context) bool {
	ctx.Send("Enter public key for new peer", telebot.RemoveKeyboard)
	return false
}

func (cmd *AddPeerCommand) addPeer(ctx telebot.Context) {
	err := cmd.ConfigManager.AddPeer(cmd.publicKey, cmd.name)
	if err != nil {
		log.Println(err)
		ctx.Send("Unexpected error occured while adding peer")
		return
	}
	log.Printf("Added new peer with public key %s and name '%s'\n", cmd.publicKey, cmd.name)
	ctx.Send("Peer was added successfully! Config below.")
	cfg, cfgStr, err := cmd.ConfigManager.GetClientConfig(cmd.publicKey)
	if err != nil {
		log.Println(err)
		ctx.Send("Unexpected error occured while trying to obtain client config for peer")
		return
	}
	configMessage := formatClientConfig(cfg, cfgStr)
	ctx.Send(configMessage, telebot.ModeMarkdownV2)
}

func (cmd *AddPeerCommand) HandleInput(ctx telebot.Context) bool {
	responseText := strings.TrimSpace(ctx.Text())
	if responseText == "" {
		return false
	}
	if cmd.publicKey == "" {
		// Handle public key input
		cmd.publicKey = responseText
		ctx.Send("Enter peer name")
		return false
	} else if cmd.name == "" {
		// Handle name input
		cmd.name = responseText
		sendConfirmation(fmt.Sprintf(addConfirmation, cmd.publicKey, cmd.name), ctx)
		return false
	} else {
		// Handle confirmaton
		switch strings.ToLower(responseText) {
		case "yes":
			cmd.addPeer(ctx)
			return true
		case "no":
			return true
		default:
			ctx.Send("Please answer 'Yes' or 'No'")
			return false
		}
	}
}
