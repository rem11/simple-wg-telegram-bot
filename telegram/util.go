package telegram

import (
	"fmt"
	"strings"

	"github.com/rem11/simple-wg-telegram-bot/wireguard"
	"gopkg.in/telebot.v3"
)

const clientConfig = "*Interface*\n" +
	"Address: `%s`\n" +
	"DNS: `%s`\n" +
	"\n" +
	"*Peer*\n" +
	"Public key: `%s`\n" +
	"Allowed IPs: `%s`\n" +
	"Endpoint: `%s`\n" +
	"\n" +
	"*Config template*\n" +
	"```\n%s\n```"

const peerLine = "%d - %s %s\n"

func formatClientConfig(cfg *wireguard.ClientConfig, cfgStr string) string {
	return fmt.Sprintf(
		clientConfig,
		cfg.Interface.Address,
		cfg.Interface.DNS,
		cfg.Peer.PublicKey,
		cfg.Peer.AllowedIPs,
		cfg.Peer.Endpoint,
		cfgStr,
	)
}

func formatPeerList(peers []wireguard.Peer) string {
	builder := strings.Builder{}
	for i, peer := range peers {
		builder.WriteString(fmt.Sprintf(peerLine, i, peer.PublicKey, peer.Name))
	}
	return builder.String()
}

func sendConfirmation(str string, ctx telebot.Context) {
	reply := ctx.Bot().NewMarkup()
	reply.Reply(reply.Row(reply.Text("Yes"), reply.Text("No")))
	reply.OneTimeKeyboard = true
	ctx.Send(str, &telebot.SendOptions{
		ReplyMarkup: reply,
		ParseMode:   telebot.ModeMarkdownV2,
	})

}
