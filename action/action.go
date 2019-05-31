package action

import (
	"fmt"
	"strings"
)

type Action struct {
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

func (a Action) isValidParamName(name string) bool {
	for _, n := range a.ParamNames {
		if n == name {
			return true
		}
	}
	return false
}

func (a Action) ParseParams(text string) ([]Param, error) {
	if len(text) == 0 {
		return nil, nil
	}

	var pp []Param
	ss := strings.Split(text, " ")
	for len(ss) > 0 {
		s := ss[0]
		var found bool
		for _, name := range a.ParamNames {
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

func (a Action) HasPermission(channelName, userName string) bool {
	if len(a.ChannelNames) > 0 {
		var found bool
		for _, n := range a.ChannelNames {
			if n == channelName {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(a.UserNames) > 0 {
		var found bool
		for _, n := range a.UserNames {
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
