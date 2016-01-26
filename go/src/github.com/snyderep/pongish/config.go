package main

import (
	"gopkg.in/gcfg.v1"
)

// Settings contains application config values.
type Settings struct {
	Server struct {
		Address       string
		StaticPrefix  string
		StaticRoot    string
		TemplateRoot  string
		WsCheckOrigin bool
	}
	Client struct {
		WebsocketGameEndpoint string
	}
}

func loadSettings(settingsFile string) (Settings, error) {
	s := Settings{}
	err := gcfg.ReadFileInto(&s, settingsFile)

	return s, err
}
