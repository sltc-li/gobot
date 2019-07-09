package handlers

import "github.com/li-go/gobot/gobot"

var (
	All = []gobot.Handler{
		helpHandler,
		lunchHandler,
		lookupHandler,
		psHandler,
		killHandler,
	}
)
