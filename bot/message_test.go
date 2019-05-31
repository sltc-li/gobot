package bot

import (
	"reflect"
	"testing"
)

func TestMessageParser_Parse(t *testing.T) {
	type fields struct {
		replyPrefix string
	}
	type args struct {
		msg       string
		channelID string
		userID    string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Message
	}{
		{
			name:   "direct message",
			fields: fields{replyPrefix: "<@PREFIX>"},
			args:   args{userID: "U123", channelID: "D123", msg: "hello"},
			want:   &Message{Type: DirectMessage, Text: "hello", ChannelID: "D123", UserID: "U123"},
		},
		{
			name:   "reply to",
			fields: fields{replyPrefix: "<@PREFIX>"},
			args:   args{userID: "U123", channelID: "X123", msg: "<@PREFIX> hello"},
			want:   &Message{Type: ReplyTo, Text: "hello", ChannelID: "X123", UserID: "U123"},
		},
		{
			name:   "listen to",
			fields: fields{replyPrefix: "<@PREFIX>"},
			args:   args{userID: "U123", channelID: "X123", msg: "hello"},
			want:   &Message{Type: ListenTo, Text: "hello", ChannelID: "X123", UserID: "U123"},
		},
		{
			name:   "clean message",
			fields: fields{replyPrefix: "<@PREFIX>"},
			args:   args{userID: "U123", channelID: "X123", msg: "  hello   world    "},
			want:   &Message{Type: ListenTo, Text: "hello world", ChannelID: "X123", UserID: "U123"},
		},
		{
			name:   "clean message - fullwidth",
			fields: fields{replyPrefix: "<@PREFIX>"},
			args:   args{userID: "U123", channelID: "X123", msg: "  　hello  　　 world  　　　  "},
			want:   &Message{Type: ListenTo, Text: "hello world", ChannelID: "X123", UserID: "U123"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := &MessageParser{
				replyPrefix: tt.fields.replyPrefix,
			}
			if got := parser.Parse(tt.args.msg, tt.args.channelID, tt.args.userID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MessageParser.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
