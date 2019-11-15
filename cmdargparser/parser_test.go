package cmdargparser

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    []Param
		wantErr bool
	}{
		{
			name:    "error - invalid pattern",
			arg:     "aaa",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "no input",
			arg:     "",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "empty value",
			arg:     "--aaa",
			want:    []Param{{Name: "aaa"}},
			wantErr: false,
		},
		{
			name:    "has value",
			arg:     "--aaa=bbb",
			want:    []Param{{Name: "aaa", Value: "bbb"}},
			wantErr: false,
		},
		{
			name:    "multiple params",
			arg:     "--aaa=bbb --ccc=ddd",
			want:    []Param{{Name: "aaa", Value: "bbb"}, {Name: "ccc", Value: "ddd"}},
			wantErr: false,
		},
		{
			name:    "multiple params - space separator",
			arg:     "--aaa bbb --ccc ddd",
			want:    []Param{{Name: "aaa", Value: "bbb"}, {Name: "ccc", Value: "ddd"}},
			wantErr: false,
		},
		{
			name:    "quoted value",
			arg:     `--aaa "bbb" --ccc "ddd eee"`,
			want:    []Param{{Name: "aaa", Value: "bbb"}, {Name: "ccc", Value: "ddd eee"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Command.ParseParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Command.ParseParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tokenize(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    []string
		wantErr bool
	}{
		{
			name:    "space separator",
			arg:     "--aaa bbb",
			want:    []string{"--aaa", "bbb"},
			wantErr: false,
		},
		{
			name:    "escaped space",
			arg:     `--aaa bbb --ccc ee\ ff`,
			want:    []string{"--aaa", "bbb", "--ccc", "ee ff"},
			wantErr: false,
		},
		{
			name:    "normal",
			arg:     `--aaa bbb --ccc ee\ ff --gg "he says \"I'm hungry.\""`,
			want:    []string{"--aaa", "bbb", "--ccc", "ee ff", "--gg", `he says "I'm hungry."`},
			wantErr: false,
		},
		{
			name:    "unpaired quote",
			arg:     `--aaa bbb --ccc ee\ ff --gg "he says \"I'm hungry.\"`,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "include equal operator",
			arg:     `--aaa=bbb --ccc=ee\ ff --gg="he says \"I'm hungry.\""`,
			want:    []string{"--aaa=bbb", "--ccc=ee ff", `--gg=he says "I'm hungry."`},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tokenize(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("tokenize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tokenize() got = %v, want %v", got, tt.want)
			}
		})
	}
}
