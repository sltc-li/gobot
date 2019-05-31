package action

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type Actions []Action

func (actions Actions) Match(text string) (*Action, string) {
	var ac *Action
	for _, a := range actions {
		if strings.HasPrefix(text, a.Name) {
			ac = &a
			break
		}
	}
	if ac == nil {
		return nil, ""
	}
	if len(text) == len(ac.Name) {
		return ac, ""
	}
	text = text[len(ac.Name):]
	if text[0] != ' ' {
		return nil, ""
	}
	return ac, text[1:]
}

func (actions Actions) Help() string {
	h := []string{"```", "available commands:"}
	for _, a := range actions {
		ss := []string{"  *", a.Name}
		for _, p := range a.ParamNames {
			ss = append(ss, fmt.Sprintf("[--%s=<%s>]", p, p))
		}
		h = append(h, strings.Join(ss, " "))
	}
	h = append(h, "```")
	return strings.Join(h, "\n")
}

func Parse(r io.Reader) (Actions, error) {
	var v struct{ Actions Actions }
	if err := json.NewDecoder(r).Decode(&v); err != nil {
		return nil, err
	}
	return v.Actions, nil
}
