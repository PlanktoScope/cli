package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

var app = &cli.App{
	Name: "planktoscope-cli",
	// TODO: see if there's a way to get the version from a build tag, so that we don't have to update
	// this manually
	Version: "v0.0.1",
	Usage:   "Command-line tool to operate and manage PlanktoScopes",
	Commands: []*cli.Command{
		devCmd,
	},
	Suggest: true,
}

// dev

var devCmd = &cli.Command{
	Name:    "dev",
	Aliases: []string{"device"},
	Usage:   "Interfaces with an individual PlanktoScope device",
	Subcommands: []*cli.Command{
		devHALCmd,
		devCtlCmd,
		devProcCmd,
	},
}

var devHALCmd = &cli.Command{
	Name:    "hal",
	Aliases: []string{"hardware-abstraction-layer"},
	Usage:   "Interfaces with a PlanktoScope device's hardware abstraction layer API",
	Subcommands: []*cli.Command{
		{
			Name:    "listen",
			Aliases: []string{"l"},
			Usage:   "Listens to and prints all messages exchanged over the API",
			Action:  devHALListenAction,
		},
	},
}

var devCtlCmd = &cli.Command{
	Name:    "ctl",
	Aliases: []string{"control"},
	Usage:   "Interfaces with a PlanktoScope device's controller API",
	Subcommands: []*cli.Command{
		{
			Name:    "listen",
			Aliases: []string{"l"},
			Usage:   "Listens to and prints all messages exchanged over the API",
			Action:  devCtlListenAction,
		},
	},
}

var devProcCmd = &cli.Command{
	Name:    "proc",
	Aliases: []string{"processing"},
	Usage:   "Interfaces with a PlanktoScope device's data processing API",
	Subcommands: []*cli.Command{
		{
			Name:    "listen",
			Aliases: []string{"l"},
			Usage:   "Listens to and prints all messages exchanged over the API",
			Action:  devProcListenAction,
		},
	},
}
