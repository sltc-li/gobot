package gobot

type Handler struct {
	Name         string
	Help         string
	NeedsMention bool
	Handleable   func(bot Bot, msg Message) bool
	Handle       func(bot Bot, msg Message) error
}

func (h Handler) IsValid() bool {
	return len(h.Name) > 0 && len(h.Help) > 0 && h.Handleable != nil && h.Handle != nil
}
