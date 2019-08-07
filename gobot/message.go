package gobot

import (
	"regexp"
	"strings"
)

//go:generate stringer -type=MsgType
type MsgType int

const (
	ListenTo MsgType = iota
	ReplyTo
	DirectMessage
)

var (
	space = regexp.MustCompile(`[\s　]+`)
	url   = regexp.MustCompile(`<(https?://[^>]+)>`)
)

type Message struct {
	Type MsgType
	Text string

	ChannelID string
	UserID    string
}

type MessageParser struct {
	replyPrefix string
}

func simplify(text string) string {
	text = space.ReplaceAllString(text, " ")
	for _, m := range url.FindAllStringSubmatch(text, -1) {
		text = strings.ReplaceAll(text, m[0], m[1])
	}
	return strings.Trim(text, " ")
}

func (parser *MessageParser) Parse(msg, channelID, userID string) Message {
	msg = simplify(msg)

	if channelID[0] == 'D' {
		return Message{Type: DirectMessage, Text: msg, ChannelID: channelID, UserID: userID}
	}
	if strings.HasPrefix(msg, parser.replyPrefix) {
		text := msg[len(parser.replyPrefix):]
		if len(text) > 0 {
			text = text[1:]
		}
		return Message{Type: ReplyTo, Text: text, ChannelID: channelID, UserID: userID}
	}
	return Message{Type: ListenTo, Text: msg, ChannelID: channelID, UserID: userID}
}

func NewMessageParser(botUserID string) *MessageParser {
	return &MessageParser{replyPrefix: "<@" + botUserID + ">"}
}
