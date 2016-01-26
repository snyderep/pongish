package main

import (
	"github.com/codegangsta/cli"
	"github.com/snyderep/pongish/server"
	"log"
	"os"
	"path"
	"path/filepath"
)

const flagConfig string = "config"

func run(c *cli.Context) {
	settings, err := loadSettings(c.String(flagConfig))
	if err != nil {
		log.Fatal(err)
	}

	s := &server.Server{
		Address: settings.Server.Address,
		Provider: server.NewPongishHandlerProvider(
			server.NewNormalTemplateRenderer(settings.Server.TemplateRoot),
			settings.Client.WebsocketGameEndpoint,
			settings.Server.WsCheckOrigin),
		StaticPrefix: settings.Server.StaticPrefix,
		StaticRoot:   settings.Server.StaticRoot}

	if err := s.Listen(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	app := cli.NewApp()
	app.Name = "pongish"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: path.Join(dir, "pongish.config.toml"),
			Usage: "configuration file",
		},
	}
	app.Action = run
	app.Run(os.Args)
}
