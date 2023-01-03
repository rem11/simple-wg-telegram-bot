package telegram

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/rem11/simple-wg-telegram-bot/wireguard"
	"gopkg.in/telebot.v3"
)

const removeConfirmation = `Are you sure that you want to remove peer?
Public key: %s
Name: %s`

type RemovePeerCommand struct {
	*wireguard.ConfigManager
	peers        []wireguard.Peer
	indexEntered bool
	index        int
}

func (cmd *RemovePeerCommand) Start(ctx telebot.Context) bool {
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
	ctx.Send(peerListStr+"\nEnter an index of peer to remove", telebot.RemoveKeyboard)
	return false
}

func (cmd *RemovePeerCommand) removePeer(ctx telebot.Context) {
	err := cmd.ConfigManager.RemovePeer(cmd.peers[cmd.index].PublicKey)
	if err != nil {
		ctx.Send("Unexpected error occured while removing peer")
		log.Println(err)
		return
	}
	ctx.Send("Peer was removed successfully!", telebot.RemoveKeyboard)
	log.Printf("Removed peer with public key %s and name '%s'\n",
		cmd.peers[cmd.index].PublicKey,
		cmd.peers[cmd.index].Name,
	)
}

func (cmd *RemovePeerCommand) HandleInput(ctx telebot.Context) bool {
	responseText := strings.TrimSpace(ctx.Text())
	if responseText == "" {
		return false
	}
	if !cmd.indexEntered {
		// Handle index
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
		cmd.indexEntered = true
		sendConfirmation(fmt.Sprintf(removeConfirmation, cmd.peers[index].PublicKey, cmd.peers[index].Name), ctx)
		return false
	} else {
		// Handle confirmaton
		switch strings.ToLower(responseText) {
		case "yes":
			cmd.removePeer(ctx)
			return true
		case "no":
			return true
		default:
			ctx.Send("Please answer 'Yes' or 'No'")
			return false
		}
	}
}
