package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/nlopes/slack"

	"github.com/li-go/gobot/action"
	"github.com/li-go/gobot/bot"
)

var (
	slackToken = os.Getenv("SLACK_TOKEN")
	rtm        = slack.New(slackToken).NewRTM()
	actions    action.Actions
	errMsgFmt  = "```\nerror:\n  %s\n```\n:thinking_face:"
	channels   = make(map[string]string)
	users      = make(map[string]string)
)

func usage() {
	executable := os.Args[0][strings.LastIndex(os.Args[0], "/")+1 : len(os.Args[0])]
	fmt.Println("Usage: \n\t" + executable + " bot.json")
	os.Exit(1)
}

func main() {
	if len(os.Args) != 2 {
		usage()
		return
	}

	reader, err := os.Open(os.Args[1])
	if err != nil {
		usage()
		return
	}

	actions, err = action.Parse(reader)
	if err != nil {
		usage()
		return
	}

	go rtm.ManageConnection()

	log.Println("start receiving message...")

	var msgParser *bot.MessageParser
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			msgParser = bot.NewMessageParser(ev.Info.User.ID)
		case *slack.MessageEvent:
			if err := loadChannel(ev.Channel); err != nil {
				log.Print(err)
				continue
			}
			if err := loadUser(ev.User); err != nil {
				log.Print(err)
				continue
			}

			// handle message
			if msgParser == nil {
				continue
			}
			msg := msgParser.Parse(ev.Text, ev.Channel, ev.User)
			handleMsg(msg)
		}
	}
}

func loadChannel(channelID string) error {
	c, err := rtm.GetConversationInfo(channelID, false)
	if err != nil {
		return fmt.Errorf("fail to get connversation: %v", err)
	}
	channels[channelID] = "#" + c.Name
	if c.IsIM {
		channels[channelID] = "<direct message>"
	}
	return nil
}

func loadUser(userID string) error {
	user, err := rtm.GetUserInfo(userID)
	if err != nil {
		return fmt.Errorf("fail to get user: %v", err)
	}
	users[userID] = "@" + user.Profile.DisplayName
	return nil
}

func handleMsg(msg *bot.Message) {
	if msg.Text == "help?" {
		rtm.SendMessage(rtm.NewOutgoingMessage(actions.Help(), msg.ChannelID))
		return
	}

	if msg.Text == "cancel" {
		// TODO: cancel action executed by the same user in the same channel
	}

	if msg.Type == bot.ReplyTo || msg.Type == bot.DirectMessage {
		ac, paramString := actions.Match(msg.Text)
		if ac == nil {
			rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf(errMsgFmt, "unknown command"), msg.ChannelID))
			return
		}
		params, err := ac.ParseParams(paramString)
		if err != nil {
			rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf(errMsgFmt, err.Error()), msg.ChannelID))
			return
		}

		if !ac.HasPermission(channels[msg.ChannelID], users[msg.UserID]) {
			rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf(errMsgFmt, "you are not allowed to do that"), msg.ChannelID))
			return
		}

		executor, err := action.NewExecutor(*ac, params)
		if err != nil {
			log.Printf("fail to create executor: %v", err)
			return
		}
		defer executor.Close()

		// post_slack hook
		go func(e *action.Executor, msg *bot.Message) {
			for {
				slackMsg, ok := e.NextSlackMessage()
				if !ok {
					break
				}
				rtm.SendMessage(rtm.NewOutgoingMessage(slackMsg, msg.ChannelID))
			}
		}(executor, msg)

		// error message hook
		go func(e *action.Executor, ac *action.Action) {
			for {
				errMsg, ok := e.NextErrorMessage()
				if !ok {
					break
				}
				if len(ac.ErrChannelID) > 0 {
					rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf(errMsgFmt, errMsg), ac.ErrChannelID))
				}
			}
		}(executor, ac)

		// execute
		log.Printf("%s is executing `%s` in %s", users[msg.UserID], executor.Command(), channels[msg.ChannelID])
		if err := executor.Exec(); err != nil {
			rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf("<@%s> *failed* - `%s` :see_no_evil:", msg.UserID, msg.Text), ac.ErrChannelID))
			log.Printf("faild: %v", err)
			return
		}
		rtm.SendMessage(rtm.NewOutgoingMessage(fmt.Sprintf("<@%s> *succeeded* - `%s` :open_mouth:", msg.UserID, msg.Text), ac.ErrChannelID))
		log.Printf("succeeded")
	}
}
