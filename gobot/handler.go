package gobot

type Handler struct {
	Name       string
	Help       func(botUser string) string
	Handleable func(bot Bot, msg Message) bool
	Handle     func(bot Bot, msg Message) error
}

func (h Handler) IsValid() bool {
	return len(h.Name) > 0 && h.Help != nil && h.Handleable != nil && h.Handle != nil
}
