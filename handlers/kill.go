package handlers

import (
	"regexp"
	"strconv"

	"github.com/li-go/gobot/configurablecommand"
	"github.com/li-go/gobot/gobot"
)

var (
	killPattern = regexp.MustCompile(`^kill (\d+)$`)
)

var killHandler = gobot.Handler{
	Name:         "kill",
	Help:         "kill %d - kill running/pending command (you can use `ps` to get command id)",
	NeedsMention: true,
	Handleable: func(bot gobot.Bot, msg gobot.Message) bool {
		return killPattern.MatchString(msg.Text)
	},
	Handle: func(bot gobot.Bot, msg gobot.Message) error {
		id, _ := strconv.Atoi(killPattern.FindStringSubmatch(msg.Text)[1])
		task, err := configurablecommand.FindTask(id)
		if err != nil {
			return err
		}
		if err := task.Kill(msg.UserID); err != nil {
			return err
		}
		bot.SendMessage("command#"+strconv.Itoa(id)+" killed!", msg.ChannelID)
		return psHandler.Handle(bot, msg)
	},
}
