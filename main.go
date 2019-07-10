package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/yaml.v2"

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
	flag.StringVar(&commandsCfg, "c", "", "commands config in yaml format")
	flag.Parse()

	var commands []configurablecommand.Command
	if len(commandsCfg) > 0 {
		file, err := os.Open(commandsCfg)
		if err != nil {
			usage(err)
		}
		err = yaml.NewDecoder(file).Decode(&commands)
		if err != nil {
			usage(err)
		}
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

	// load pending tasks
	configurablecommand.LoadPendingTasks(bot)

	// wait signal
	signCh := make(chan os.Signal)
	signal.Notify(signCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signCh
		configurablecommand.StopAll()
		bot.Stop()
		os.Exit(1)
	}()

	// start
	bot.Start()
}
