package handlers

import (
	"encoding/json"
	"github.com/li-go/gobot/gobot"
	"regexp"
)

var (
	lookupPattern = regexp.MustCompile(`^lookup: <@(\w+)>$`)
)

var LookupHandler = gobot.Handler{
	Name: "lookup",
	Help: func(botUser string) string {
		return "lookup: @someone"
	},
	Handleable: func(bot gobot.Bot, msg gobot.Message) bool {
		return lookupPattern.MatchString(msg.Text)
	},
	Handle: func(bot gobot.Bot, msg gobot.Message) error {
		userID := lookupPattern.FindStringSubmatch(msg.Text)[1]
		user, err := bot.GetRTM().GetUserInfo(userID)
		if err != nil {
			return err
		}

		simpleUser := struct{ ID, Name, DisplayName string }{
			ID:          user.ID,
			Name:        user.Name,
			DisplayName: user.Profile.DisplayName,
		}
		buf, err := json.MarshalIndent(simpleUser, "", "  ")
		if err != nil {
			return err
		}

		bot.SendMessage("```\n"+string(buf)+"\n```", msg.ChannelID)
		return nil
	},
}
