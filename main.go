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
func check_ack(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "Go-bean"
	app.Version = "0.0.1"

	app.Usage = "Deploy your go services to AWS ElasticBeanstalk wrapped in docker containers"

	force_command := false
	release_version := ""

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "force, f",
			Usage:       "Force the operation (ignoring warnings)",
			Destination: &force_command,
		},
		cli.StringFlag{
			Name:        "release-version, rv",
			Usage:       "The version of the code to release or deploy",
			Destination: &release_version,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "init",
			Usage: "Initialize a directory as go-bean enabled.",
			Action: func(c *cli.Context) {

				err := Command_Init(force_command)
				check_ack(err)
			},
		},
		{
			Name:  "config",
			Usage: "Show the current configuration",
			Action: func(c *cli.Context) {

				err := Command_Config()
				check_ack(err)
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
			Name:  "snapshot",
			Usage: "Get snapshot version, build, deploy to ECR, create new application version on EB",
			Action: func(c *cli.Context) {
				conf, config_err := LoadConfig()
				check_ack(config_err)
				check_ack(PerformSnapshot(conf))
			},
		},
		{
			Name:  "eb-status",
			Usage: "Display the EB status for the assigned environment",
			Action: func(c *cli.Context) {
				conf, config_err := LoadConfig()
				check_ack(config_err)

				state, err := CurrentEnvState(conf.EnvName)
				check_ack(err)
				println("State: " + state.Status)
				println("Version: " + state.VersionLabel)
			},
		},
		{
			Name:  "release",
			Usage: "Bump version, build, deploy to ECR, create new application version on EB",
			Action: func(c *cli.Context) {
				conf, config_err := LoadConfig()
				check_ack(config_err)
				check_ack(PerformRelease(conf, release_version))
			},
		},
	}

	app.Run(os.Args)
}
