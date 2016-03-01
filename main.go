package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"os"
)

// You can do the following things:
// init (inits the directory, creating the .go-bean file
// release (bumps the version (you can add one manually)), creates the docker container, and uploads to ECR, creates version for application)
// deploy (deploys the latest version (or specific) to aws EB environment (specify)

func main() {
	app := cli.NewApp()
	app.Name = "Go-bean"
	app.Version = "0.0.1"

	app.Usage = "Deploy your go services to AWS ElasticBeanstalk wrapped in docker containers"

	force_command := false
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "force, f",
			Usage:       "Force the operation (ignoring warnings)",
			Destination: &force_command,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "init",
			Usage: "Initialize a directory as go-bean enabled.",
			Action: func(c *cli.Context) {

				err := Command_Init(force_command)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			},
		},
		{
			Name:  "config",
			Usage: "Show the current configuration",
			Action: func(c *cli.Context) {

				err := Command_Config()
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			},
		},
		{
			Name:  "deploy",
			Usage: "Deploy a version of the app to an ElasticBeanstalk environment",
			Action: func(c *cli.Context) {
				println("Deploy!")
			},
		},
		{
			Name:  "release",
			Usage: "Bump version, build, deploy to ECR, create new application version on EB",
			Action: func(c *cli.Context) {
				println("releasing!")
				if force_command {
					println("Force")
				} else {
					println("Not")
				}
			},
		},
	}

	app.Run(os.Args)
}
