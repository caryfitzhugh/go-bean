package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"os"
)

const tag_prefix = "go-bean"

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
				// Find the version (get the git short hash + the version currently)
				conf, config_err := LoadConfig()
				check_ack(config_err)

				tag, err_tag := GetGitTag(tag_prefix)
				check_ack(err_tag)
				sha, err_ref := GetGitRef()
				check_ack(err_ref)
				if tag == sha {
					tag = "initial"
				}

				snapshot_version := tag + "-" + sha
				println("Creating snapshot version: " + snapshot_version)
				println("Tagging git...")
				check_ack(GitTagCurrent(tag_prefix, snapshot_version))

				println("Building static linux executable")
				build_err := BuildBinary(conf.ProgramName)
				check_ack(build_err)

				println("Building docker container")
				docker_build_err := BuildDockerImage(conf.ProgramName, conf.ProgramPort, snapshot_version)
				check_ack(docker_build_err)

				println("Logging into ECR")
				check_ack(LoginToECR())

				// Now push it to ECR
				println("Pushing image to container repository")
				docker_push_err := PushToRepository(conf.ProgramName, snapshot_version, conf.DockerHost)
				check_ack(docker_push_err)

				println("Creating Application Version")
				if app_ver, err := CreateApplicationVersion(conf.AppName, snapshot_version, conf.ProgramName, conf.DockerHost, conf.ProgramPort, conf.S3Bucket); err != nil {
					check_ack(err)
				} else {
					println("Updating env: " + conf.EnvName)
					check_ack(UpdateEBEnvironment(conf.EnvName, app_ver))

					println("Waiting for update to complete.")
					if err := WaitForEBToBeReady(conf.EnvName, app_ver); err != nil {
						println("Error? What's up!?")
					} else {
						println("Done!")
					}
				}
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
				// We need to find out what version to release.
				// Someone needs to increment the version # automatically, or get it from the command line.

				println("Release")
			},
		},
	}

	app.Run(os.Args)
}
