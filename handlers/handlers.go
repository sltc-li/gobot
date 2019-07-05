package handlers

import "github.com/li-go/gobot/gobot"

var (
	All = []gobot.Handler{
		helpHandler,
		lookupHandler,
		psHandler,
		killHandler,
	}
)
