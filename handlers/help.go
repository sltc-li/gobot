package handlers

import "github.com/li-go/gobot/gobot"

var helpHandler = gobot.Handler{
	Name: "help",
	Help: "help? - print help information",
	Handleable: func(bot gobot.Bot, msg gobot.Message) bool {
		return msg.Text == "help?"
	},
	Handle: func(bot gobot.Bot, msg gobot.Message) error {
		bot.SendMessage(bot.Help(), msg.ChannelID)
		return nil
	},
}
