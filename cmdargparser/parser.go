package cmdargparser

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	spaceSeparatedArgNamePattern = regexp.MustCompile(`^--(\w+)$`)
	equalJoinedArgNamePattern    = regexp.MustCompile(`^--(\w+)=(.+)$`)

	ErrUnpairedQuote     = errors.New("unpaired quote")
	ErrInvalidArgPattern = errors.New("invalid argument pattern")
)

type Param struct {
	Name  string
	Value string
}

func Parse(text string) ([]Param, error) {
	tokens, err := tokenize(text)
	if err != nil {
		return nil, err
	}

	var pp []Param
	for len(tokens) > 0 {
		token := tokens[0]
		tokens = tokens[1:]
		matched := spaceSeparatedArgNamePattern.FindStringSubmatch(token)
		if len(matched) > 0 {
			param := Param{Name: matched[1]}
			if len(tokens) > 0 && !spaceSeparatedArgNamePattern.MatchString(tokens[0]) {
				param.Value = tokens[0]
				tokens = tokens[1:]
			}
			pp = append(pp, param)
			continue
		}
		matched = equalJoinedArgNamePattern.FindStringSubmatch(token)
		if len(matched) > 0 {
			pp = append(pp, Param{Name: matched[1], Value: matched[2]})
			continue
		}
		return nil, fmt.Errorf("token: %s: %w", token, ErrInvalidArgPattern)
	}
	return pp, nil
}

func tokenize(text string) ([]string, error) {
	var tokens []string
	var tmp []rune
	var inQuote bool
	for i, c := range text {
		switch c {
		case ' ':
			if i > 0 && text[i-1] == '\\' {
				tmp = append(tmp, c)
				continue
			}
			if inQuote {
				tmp = append(tmp, c)
				continue
			}
			if tmp != nil {
				tokens = append(tokens, string(tmp))
				tmp = nil
			}
		case '"':
			if i > 0 && text[i-1] == '\\' {
				tmp = append(tmp, c)
				continue
			}
			if inQuote {
				tokens = append(tokens, string(tmp))
				tmp = nil
				inQuote = false
				continue
			}
			inQuote = true
		case '\\':
		default:
			tmp = append(tmp, c)
		}
	}
	if inQuote {
		return nil, ErrUnpairedQuote
	}
	if tmp != nil {
		return append(tokens, string(tmp)), nil
	}
	return tokens, nil
}
