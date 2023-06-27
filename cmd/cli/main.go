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
	Name: "planktoscope",
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
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "api",
			Value:   defaultAPIURL,
			Usage:   "Path of the PlanktoScope's API",
			EnvVars: []string{"PLANKTOSCOPE_API"},
		},
		&cli.StringFlag{
			Name:    "instance-id",
			Aliases: []string{"id"},
			Value:   "",
			Usage:   "MQTT client instance ID of the API client",
			EnvVars: []string{"PLANKTOSCOPE_CLIENT_INSTANCE_ID"},
		},
	},
	Subcommands: []*cli.Command{
		{
			Name:   "listen",
			Usage:  "Listens to and prints all messages exchanged over the API",
			Action: devListenAction,
		},
		devHALCmd,
		devCtlCmd,
		devProcCmd,
	},
}

const defaultAPIURL = "mqtt://home.planktoscope:1883"

var devHALCmd = &cli.Command{
	Name:    "hal",
	Aliases: []string{"hardware", "hardware-abstraction-layer"},
	Usage:   "Interfaces with a PlanktoScope device's hardware abstraction layer API",
	Subcommands: []*cli.Command{
		{
			Name:   "listen",
			Usage:  "Listens to and prints all messages exchanged over the API",
			Action: devHALListenAction,
		},
	},
}

var devCtlCmd = &cli.Command{
	Name:    "ctl",
	Aliases: []string{"control"},
	Usage:   "Interfaces with a PlanktoScope device's controller API",
	Subcommands: []*cli.Command{
		{
			Name:   "listen",
			Usage:  "Listens to and prints all messages exchanged over the API",
			Action: devCtlListenAction,
		},
	},
}

var devProcCmd = &cli.Command{
	Name:    "proc",
	Aliases: []string{"processing"},
	Usage:   "Interfaces with a PlanktoScope device's data processing API",
	Subcommands: []*cli.Command{
		{
			Name:   "listen",
			Usage:  "Listens to and prints all messages exchanged over the API",
			Action: devProcListenAction,
		},
		{
			Name:   "start",
			Usage:  "Begins a data processing routine on the PlanktoScope device",
			Action: devProcStartAction,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "path",
					Value: "/home/pi/data/img",
					Usage: "Root directory on the PlanktoScope device of raw datasets to process",
				},
				&cli.Uint64Flag{
					Name:  "processing-id",
					Value: 1,
					Usage: "Unique ID of the data processing routine",
				},
				&cli.BoolFlag{
					Name:  "recurse",
					Value: true,
					Usage: "Whether to recurse into all child directories of the root directory when " +
						"identifying datasets to process",
				},
				&cli.BoolFlag{
					Name:  "force-reprocessing",
					Value: false,
					Usage: "Whether to run the processing routine on datasets for which processing results " +
						"already exist",
				},
				&cli.BoolFlag{
					Name:  "keep-objects",
					Value: true,
					Usage: "Whether to keep individual images of isolated objects in the processing results",
				},
				&cli.BoolFlag{
					Name:  "export-ecotaxa",
					Value: true,
					Usage: "Whether to export the processing results as an archive for upload to EcoTaxa",
				},
				&cli.BoolFlag{
					Name:  "await-started",
					Value: true,
					Usage: "Whether to wait for confirmation from the data processing API that the " +
						"processing routine has started before exiting",
				},
				&cli.BoolFlag{
					Name:  "await-finished",
					Value: true,
					Usage: "Whether to wait for confirmation from the data processing API that the " +
						"processing routine has finished before exiting",
				},
			},
		},
	},
}
