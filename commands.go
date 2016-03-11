package main

import (
	"errors"
	//"fmt"
	"github.com/hashicorp/go-version"
	//	"strings"
)

func PerformSnapshot(conf GoBeanConfig) error {
	// Find the version (get the git short hash + the version currently)
	cur_ver, err_tag := GetCurVer()
	check_ack(err_tag)
	sha, err_ref := GetGitRef()
	check_ack(err_ref)

	snapshot_version := cur_ver + "@" + sha

	if !GitAllCheckedIn() {
		snapshot_version += "-dirty"
	}

	println("Snapshot version:" + snapshot_version)
	DeployVersion(conf, snapshot_version)
	return nil
}

func PerformRelease(conf GoBeanConfig, release_version string) error {
	if !GitAllCheckedIn() {
		return errors.New("To release, you must have a clean working tree. Commit your changes and try again")
	}
	var new_version string

	if release_version == "" {
		// Find the version (get the git short hash + the version currently)
		cur_ver, err := GetCurVer()
		if err != nil {
			return errors.New("Could not get the current version")
		}

		var new_ver *version.Version
		new_ver, err = version.NewVersion(cur_ver)
		segments := new_ver.Segments()

		// Increment the last value
		segments[len(segments)-1] = segments[len(segments)-1] + 1

		// Convert back to a string
		new_version = new_ver.String()

	} else {
		_, err := version.NewVersion(release_version)
		if err != nil {
			return errors.New("The release you provided was not a valid SymVer release: " + release_version)
		}

		new_version = release_version
	}

	println("Releasing " + new_version)

	SaveCurVer(new_version)

	DeployVersion(conf, new_version)
	return nil
}

func DeployVersion(conf GoBeanConfig, version string) error {
	println("Building static linux executable")
	err := BuildBinary(conf.ProgramName)
	if err != nil {
		return err
	}

	println("Building docker container")
	err = BuildDockerImage(conf.ProgramName, conf.ProgramPort, version)
	if err != nil {
		return err
	}

	err = LoginToECR()
	if err != nil {
		return err
	}

	// Now push it to ECR
	println("Pushing image to container repository")
	err = PushToRepository(conf.ProgramName, version, conf.DockerHost)
	if err != nil {
		return err
	}

	println("Creating Application Version")
	var app_ver string

	app_ver, err = CreateApplicationVersion(conf.AppName, version, conf.ProgramName, conf.DockerHost, conf.ProgramPort, conf.S3Bucket)
	if err != nil {
		return err
	}

	println("Updating env: " + conf.EnvName)
	err = UpdateEBEnvironment(conf.EnvName, app_ver)
	if err != nil {
		return err
	}

	println("Waiting for update to complete.")
	err = WaitForEBToBeReady(conf.EnvName, app_ver)
	if err != nil {
		return err
	} else {
		println("Done!")
	}
	return nil
}
