package handlers

import (
	"strconv"
	"strings"

	"github.com/li-go/gobot/configurablecommand"
	"github.com/li-go/gobot/gobot"
)

var ListHandler = gobot.Handler{
	Name:         "ps",
	Help:         "ps - list running/finished commands",
	NeedsMention: true,
	Handleable: func(bot gobot.Bot, msg gobot.Message) bool {
		return msg.Text == "ps"
	},
	Handle: func(bot gobot.Bot, msg gobot.Message) error {
		tasks := configurablecommand.GetTasks()
		var tt []configurablecommand.Task
		for _, task := range tasks {
			if task.Msg.ChannelID != msg.ChannelID {
				continue
			}
			tt = append(tt, task)
		}
		var ss []string
		for _, task := range tt {
			user, err := bot.LoadUser(task.Msg.UserID)
			if err != nil {
				user = "anonymous"
			}
			s := "  * " + strconv.Itoa(task.ID) + ". (" + task.Status().String() + ") " +
				user + ": " + task.Msg.Text +
				" (time: " + task.Duration().String() + ")"
			ss = append(ss, s)
		}
		text := "```\n" + "Latest commands:\n" + strings.Join(ss, "\n") + "\n```"
		bot.SendMessage(text, msg.ChannelID)
		return nil
	},
}
