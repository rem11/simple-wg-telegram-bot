package telegram

import (
	"log"
	"strconv"
	"strings"

	"github.com/rem11/simple-wg-telegram-bot/wireguard"
	"gopkg.in/telebot.v3"
)

type ClientConfigCommand struct {
	*wireguard.ConfigManager
	peers []wireguard.Peer
	index int
}

func (cmd *ClientConfigCommand) Start(ctx telebot.Context) bool {
	peers, err := cmd.ConfigManager.ListPeers()
	if err != nil {
		ctx.Send("Unexpected error while fetching peer list")
		log.Println(err)
		return true
	}
	if len(peers) == 0 {
		ctx.Send("No peers found in configuration")
		return true
	}
	cmd.peers = peers
	peerListStr := formatPeerList(peers)
	ctx.Send(peerListStr+"\nEnter an index of peer to display its client configuration", telebot.RemoveKeyboard)
	return false
}

func (cmd *ClientConfigCommand) getClientConfig(ctx telebot.Context) {
	cfg, cfgStr, err := cmd.ConfigManager.GetClientConfig(cmd.peers[cmd.index].PublicKey)
	if err != nil {
		log.Println(err)
		ctx.Send("Unexpected error occured while trying to obtain client config for peer")
		return
	}
	configMessage := formatClientConfig(cfg, cfgStr)
	ctx.Send(configMessage, telebot.ModeMarkdownV2)
}

func (cmd *ClientConfigCommand) HandleInput(ctx telebot.Context) bool {
	responseText := strings.TrimSpace(ctx.Text())
	if responseText == "" {
		return false
	}
	index, err := strconv.Atoi(responseText)
	if err != nil {
		ctx.Send("Please enter a number")
		return false
	}
	if index >= len(cmd.peers) || index < 0 {
		ctx.Send("Index is out of range")
		return false
	}
	cmd.index = index
	cmd.getClientConfig(ctx)
	return true
}
