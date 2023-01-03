package main

import (
	"flag"
	"log"
	"time"

	"github.com/rem11/simple-wg-telegram-bot/telegram"
	"github.com/rem11/simple-wg-telegram-bot/wireguard"
	"gopkg.in/ini.v1"
)

type Config struct {
	ConfigFilePath string
	Hostname       string
	DNS            string
	UseStub        bool
	InterfaceName  string
	BotToken       string
	UserIDs        []int64
}

func readConfig(configPath string) *Config {
	cfgFile, err := ini.Load(configPath)
	if err != nil {
		log.Fatal(err)
	}

	config := &Config{}

	err = cfgFile.MapTo(config)
	if err != nil {
		log.Fatal(err)
	}

	return config
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Configuration file path")
	flag.Parse()

	if configPath == "" {
		log.Fatal("Please specify path to configuration file")
	}

	config := readConfig(configPath)

	var processManager wireguard.ProcessManagerInterface
	if config.UseStub {
		processManager = &wireguard.ProcessManagerStub{}
	} else {
		processManager = &wireguard.ProcessManager{
			InterfaceName: config.InterfaceName,
		}
	}

	configManager := &wireguard.ConfigManager{
		ConfigFilePath: config.ConfigFilePath,
		Hostname:       config.Hostname,
		DNS:            config.DNS,
		ProcessManager: processManager,
	}

	bot := telegram.Bot{
		ConfigManager:     configManager,
		CommandController: telegram.NewCommandController(),
		PollingTimeout:    30 * time.Second,
		Token:             config.BotToken,
		UserIDs:           config.UserIDs,
	}

	err := bot.Start()
	if err != nil {
		log.Fatal(err)
	}
}
