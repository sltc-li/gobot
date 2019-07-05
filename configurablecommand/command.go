package configurablecommand

import (
	"errors"
	"fmt"
	"github.com/li-go/gobot/gobot"
	"strings"
)

var (
	errMsgFmt = "```\nerror:\n  %s\n```\n:thinking_face:"

	ErrNoPermission = errors.New("no permission")
)

type Command struct {
	Name         string
	Command      string
	ParamNames   []string `json:"params"`
	LogFilename  string   `json:"log"`
	ErrChannelID string   `json:"error_channel"`
	ChannelNames []string `json:"channels"`
	UserNames    []string `json:"users"`
}

type Param struct {
	Name  string
	Value string
}

func (c Command) isValidParamName(name string) bool {
	for _, n := range c.ParamNames {
		if n == name {
			return true
		}
	}
	return false
}

func (c Command) ParseParams(text string) ([]Param, error) {
	if len(text) == 0 {
		return nil, nil
	}

	var pp []Param
	ss := strings.Split(text, " ")
	for len(ss) > 0 {
		s := ss[0]
		var found bool
		for _, name := range c.ParamNames {
			if s == "--"+name {
				value := ""
				if len(ss) > 1 {
					value = ss[1]
					ss = ss[1:]
				}
				pp = append(pp, Param{Name: name, Value: value})
				found = true
				break
			}
			if strings.HasPrefix(s, "--"+name+"=") {
				pp = append(pp, Param{Name: name, Value: s[len(name)+3:]})
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("invalid param: %s", s)
		}
		ss = ss[1:]
	}
	return pp, nil
}

func (c Command) HasPermission(channelName, userName string) bool {
	if len(c.ChannelNames) > 0 {
		var found bool
		for _, n := range c.ChannelNames {
			if n == channelName {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(c.UserNames) > 0 {
		var found bool
		for _, n := range c.UserNames {
			if n == userName {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (c Command) Match(text string) (bool, string) {
	if !strings.HasPrefix(text, c.Name) {
		return false, ""
	}
	if len(text) == len(c.Name) {
		return true, ""
	}
	text = text[len(c.Name):]
	if text[0] != ' ' {
		return false, ""
	}
	return true, text[1:]
}

func (c Command) Handler() gobot.Handler {
	return gobot.Handler{
		Name: c.Name,
		Help: func(botUser string) string {
			return botUser + " " + c.help()
		},
		Handleable: func(bot gobot.Bot, msg gobot.Message) bool {
			if msg.Type == gobot.ListenTo {
				return false
			}
			m, _ := c.Match(msg.Text)
			return m
		},
		Handle: func(bot gobot.Bot, msg gobot.Message) error {
			task := addTask(c, msg)
			err := c.run(bot, msg)
			finishTask(task, err)
			return err
		},
	}
}

func (c Command) help() string {
	ss := []string{c.Name}
	for _, p := range c.ParamNames {
		ss = append(ss, fmt.Sprintf("[--%s=<%s>]", p, p))
	}
	return strings.Join(ss, " ")
}

func (c Command) run(bot gobot.Bot, msg gobot.Message) error {
	_, paramString := c.Match(msg.Text)
	params, err := c.ParseParams(paramString)
	if err != nil {
		bot.SendMessage(fmt.Sprintf(errMsgFmt, err.Error()), msg.ChannelID)
		return err
	}

	channel, err := bot.LoadChannel(msg.ChannelID)
	if err != nil {
		return err
	}
	user, err := bot.LoadUser(msg.UserID)
	if err != nil {
		return err
	}
	if !c.HasPermission(channel, user) {
		bot.SendMessage(fmt.Sprintf(errMsgFmt, "you are not allowed to do that"), msg.ChannelID)
		return ErrNoPermission
	}

	executor, err := NewExecutor(c, params)
	if err != nil {
		return err
	}
	defer executor.Close()

	// post_slack hook
	go func(e *Executor, msg gobot.Message) {
		for {
			slackMsg, ok := e.NextSlackMessage()
			if !ok {
				break
			}
			bot.SendMessage(slackMsg, msg.ChannelID)
		}
	}(executor, msg)

	// error message hook
	go func(e *Executor, c Command) {
		for {
			errMsg, ok := e.NextErrorMessage()
			if !ok {
				break
			}
			if len(c.ErrChannelID) > 0 {
				bot.SendMessage(fmt.Sprintf(errMsgFmt, errMsg), c.ErrChannelID)
			}
		}
	}(executor, c)

	// execute
	bot.GetLogger().Printf("%s is executing `%s` in %s", user, executor.Command(), channel)
	if err := executor.Exec(); err != nil {
		if len(c.ErrChannelID) > 0 {
			bot.SendMessage(fmt.Sprintf("<@%s> *failed* - `%s` :see_no_evil:", msg.UserID, msg.Text), c.ErrChannelID)
		}
		return err
	}
	if len(c.ErrChannelID) > 0 {
		bot.SendMessage(fmt.Sprintf("<@%s> *succeeded* - `%s` :open_mouth:", msg.UserID, msg.Text), c.ErrChannelID)
	}
	bot.GetLogger().Printf("succeeded")
	return nil
}
