package main

import (
	"errors"
	"fmt"
	"github.com/hashicorp/go-version"
	"time"
	//	"strings"
)

func PerformSnapshot(conf GoBeanConfig) error {
	// Find the version (get the git short hash + the version currently)
	cur_ver, err_tag := GetCurVer()
	check_ack(err_tag)
	sha, err_ref := GetGitRef()
	check_ack(err_ref)

	snapshot_version := fmt.Sprintf("%v-%v-%v", cur_ver, sha, time.Now().Unix())

	if !GitAllCheckedIn() {
		snapshot_version += "-dirty"
	}

	println("Snapshot version:" + snapshot_version)
	return DeployVersion(conf, snapshot_version)
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

	return DeployVersion(conf, new_version)
}

func DeployVersion(conf GoBeanConfig, version string) error {
	print("Building static linux executable...")
	err := BuildBinary(conf.ProgramName)
	if err != nil {
		return err
	}
	println("Done")

	print("Building docker container...")
	err = BuildDockerImage(conf.ProgramName, conf.ProgramPort, version)
	if err != nil {
		return err
	}
	println("Done")

	print("Logging into ECR...")
	err = LoginToECR()
	if err != nil {
		return err
	}
	println("Done")

	// Now push it to ECR
	print("Pushing image to container repository...")
	err = PushToRepository(conf.ProgramName, version, conf.DockerHost)
	if err != nil {
		return err
	}
	println("Done")

	print("Creating Application Version...")
	var app_ver string

	app_ver, err = CreateApplicationVersion(conf.AppName, version, conf.ProgramName, conf.DockerHost, conf.ProgramPort, conf.S3Bucket)
	if err != nil {
		return err
	}
	println("Done")

	print("Updating env: " + conf.EnvName + "...")
	err = UpdateEBEnvironment(conf.EnvName, app_ver)
	if err != nil {
		return err
	}
	println("Done")

	print("Waiting for update to complete.")
	err = WaitForEBToBeReady(conf.EnvName, app_ver)
	if err != nil {
		return err
	} else {
		println("Done!")
	}
	return nil
}
