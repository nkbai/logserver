package main

import (
	"fmt"

	"os"

	"log"

	"gopkg.in/urfave/cli.v1"
)

func main() {
	StartMain()
}

var logdir string

//StartMain entry point of Photon app
func StartMain() {
	fmt.Printf("os.args=%q\n", os.Args)
	app := cli.NewApp()
	wd, _ := os.Getwd()
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "port",
			Usage: ` port  for listening.`,
			Value: 5000,
		},
		cli.StringFlag{
			Name:  "logdir",
			Usage: "log save dir",
			Value: wd,
		},
	}
	app.Action = mainCtx
	app.Name = "logserver"
	app.Version = "0.1"

	err := app.Run(os.Args)
	if err != nil {
		log.Printf("run err %s", err)
	}
}

func mainCtx(ctx *cli.Context) error {
	fmt.Printf("Welcom to logserver,version %s\n", ctx.App.Version)
	logdir = ctx.String("logdir")
	Start(ctx.Int("port"))
	return nil
}
