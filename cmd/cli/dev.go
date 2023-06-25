package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// hal

func devHALListenAction(c *cli.Context) error {
	fmt.Println("hello, world!")
	return nil
}

// ctl

func devCtlListenAction(c *cli.Context) error {
	fmt.Println("hello, world!")
	return nil
}

// proc

func devProcListenAction(c *cli.Context) error {
	fmt.Println("hello, world!")
	return nil
}
