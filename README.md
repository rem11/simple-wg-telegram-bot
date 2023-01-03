# About

This is a simple Telegram bot for managing Wireguard VPN configuration. It allows to add and remove peers and retrieve peer's client configuration.

The main idea was to provide a way to configure a Wireguard VPN server without exposing any configuration consoles to Internet. This bot does not generate a private keys for peers to avoid sending them over insecure medium (i.e. Telegram). So you need to create an empty configuration on your client device first, and provide your public key when adding a new peer via this bot.

This bot does not rely on any additional databases and stores all configuration in Wireguard configuration file.

# Installation

Fetch latest sources:

```
git clone https://github.com/rem11/simple-wg-telegram-bot.git
```

Navigate to repository, and invoke `go install`.

# Running

Create a configuration file, according to below example:

```
; Path to wiregurad configuration
ConfigFilePath = /etc/wireguard/wg0.conf
; VPN server hostname to be used in client configurations
Hostname = test.example.com
; DNS server to be used in client configurations
DNS = 8.8.8.8
; Use stub process manager which does not perform any actual config reloading in wireguard
UseStub = false
; Wireguard interface to reload config for
InterfaceName = wg0
; Telegram bot token
BotToken = xxx
; Telegram user IDs who allowed to use this bot
UserIDs = 111222333
```

Start a program with a path to the config file:

```
simple-wg-telegram-bot -config wg-bot.conf
```
