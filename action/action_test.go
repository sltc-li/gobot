package action

import (
	"reflect"
	"testing"
)

func TestAction_ParseParams(t *testing.T) {
	type fields struct {
		Name       string
		ParamNames []string
	}
	type args struct {
		text string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Param
		wantErr bool
	}{
		{
			name:    "error - invalid format",
			fields:  fields{ParamNames: []string{}},
			args:    args{text: "aaa"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "error - unknown param name",
			fields:  fields{ParamNames: []string{}},
			args:    args{text: "--aaa"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "no input",
			fields:  fields{ParamNames: []string{"aaa"}},
			args:    args{text: ""},
			want:    nil,
			wantErr: false,
		},
		{
			name:    "no value",
			fields:  fields{ParamNames: []string{"aaa"}},
			args:    args{text: "--aaa"},
			want:    []Param{{Name: "aaa"}},
			wantErr: false,
		},
		{
			name:    "has value",
			fields:  fields{ParamNames: []string{"aaa"}},
			args:    args{text: "--aaa=bbb"},
			want:    []Param{{Name: "aaa", Value: "bbb"}},
			wantErr: false,
		},
		{
			name:    "multiple params",
			fields:  fields{ParamNames: []string{"aaa", "ccc"}},
			args:    args{text: "--aaa=bbb --ccc=ddd"},
			want:    []Param{{Name: "aaa", Value: "bbb"}, {Name: "ccc", Value: "ddd"}},
			wantErr: false,
		},
		{
			name:    "multiple params - space separator",
			fields:  fields{ParamNames: []string{"aaa", "ccc"}},
			args:    args{text: "--aaa bbb --ccc ddd"},
			want:    []Param{{Name: "aaa", Value: "bbb"}, {Name: "ccc", Value: "ddd"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Action{
				ParamNames: tt.fields.ParamNames,
			}
			got, err := a.ParseParams(tt.args.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("Action.ParseParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Action.ParseParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestActions_Match(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name    string
		actions Actions
		args    args
		want    *Action
		want1   string
	}{
		{
			name:    "not found",
			actions: Actions{Action{Name: "aaa"}},
			args:    args{text: "xxx"},
			want:    nil,
			want1:   "",
		},
		{
			name:    "not found - not following space",
			actions: Actions{Action{Name: "aaa"}},
			args:    args{text: "aaabbb"},
			want:    nil,
			want1:   "",
		},
		{
			name:    "found",
			actions: Actions{Action{Name: "aaa"}},
			args:    args{text: "aaa"},
			want:    &Action{Name: "aaa"},
			want1:   "",
		},
		{
			name:    "found - with follow text",
			actions: Actions{Action{Name: "aaa"}},
			args:    args{text: "aaa bbb ccc"},
			want:    &Action{Name: "aaa"},
			want1:   "bbb ccc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.actions.Match(tt.args.text)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Actions.Match() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Actions.Match() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
