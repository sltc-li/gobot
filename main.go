package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/li-go/gobot/configurablecommand"
	"github.com/li-go/gobot/gobot"
	"github.com/li-go/gobot/handlers"
)

var (
	commandsCfg string
)

func usage(err error) {
	println("error: " + err.Error())
	flag.Usage()
	os.Exit(2)
}

func main() {
	flag.StringVar(&commandsCfg, "commands-conf", "", "commands configuration in json format")
	flag.Parse()

	var commands []configurablecommand.Command
	if len(commandsCfg) > 0 {
		reader, err := os.Open(commandsCfg)
		if err != nil {
			usage(err)
		}
		var v struct{ Commands []configurablecommand.Command }
		err = json.NewDecoder(reader).Decode(&v)
		if err != nil {
			usage(err)
		}
		commands = v.Commands
	}

	logger := log.New(os.Stdout, "bot: ", log.LstdFlags)
	bot, err := gobot.New(os.Getenv("SLACK_TOKEN"), logger)
	if err != nil {
		usage(err)
	}

	// register defined handlers
	for _, h := range handlers.All {
		if err := bot.RegisterHandler(h); err != nil {
			usage(err)
		}
	}

	// register configurable command handlers
	for _, c := range commands {
		if err := bot.RegisterHandler(c.Handler()); err != nil {
			usage(err)
		}
	}

	// start
	bot.Start()
}
