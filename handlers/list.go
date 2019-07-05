package handlers

import (
	"github.com/li-go/gobot/configurablecommand"
	"github.com/li-go/gobot/gobot"
	"sort"
	"strings"
)

var (
	listMaxTasks = 10
)

var ListHandler = gobot.Handler{
	Name:         "list",
	Help:         "list - list running/finished commands",
	NeedsMention: true,
	Handleable: func(bot gobot.Bot, msg gobot.Message) bool {
		return msg.Text == "list"
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
		sort.Slice(tt, func(i, j int) bool {
			return configurablecommand.LessFunc(tt[i], tt[j])
		})
		if len(tt) > listMaxTasks {
			tt = tt[len(tt)-listMaxTasks:]
		}
		var ss []string
		for _, task := range tt {
			user, err := bot.LoadUser(task.Msg.UserID)
			if err != nil {
				user = "unknown"
			}
			s := "  * " + user
			if task.EndAt == nil {
				s += " is executing"
			} else {
				s += " executed"
			}
			s += " `" + task.Msg.Text + "`"
			s += " (" + task.StartAt.Format("01/02 15:04") + " ~"
			if task.EndAt != nil {
				s += " " + task.EndAt.Format("01/02 15:04")
			}
			s += ")"
			if task.Err != nil {
				s += " (error: " + task.Err.Error() + ")"
			}
			ss = append(ss, s)
		}
		text := "```\n" + "Latest commands:\n" + strings.Join(ss, "\n") + "\n```"
		bot.SendMessage(text, msg.ChannelID)
		return nil
	},
}
