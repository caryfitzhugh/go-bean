package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

func GitTagCurrent(prefix string, ver string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	} else {
		return GitTagCurrentInDir(prefix, ver, pwd)
	}
}
func GitTagCurrentInDir(prefix string, ver string, working_dir string) error {
	cmdName := "git"
	cmdArgs := []string{"tag", "-f", prefix + "-" + ver}
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = working_dir
	if err := cmd.Run(); err != nil {
		return errors.New("There was an error running git tag command")
	}
	return nil
}

func GetGitRef() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	} else {
		return GetGitRefFromDir(pwd)
	}
}

func GetGitRefFromDir(working_dir string) (string, error) {
	var (
		cmdOut []byte
		err    error
	)

	cmdName := "git"
	cmdArgs := []string{"rev-parse", "--short", "HEAD"}
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = working_dir

	if cmdOut, err = cmd.Output(); err != nil {
		return "", errors.New("There was an error running git rev-parse command")
	}
	sha := strings.TrimSpace(string(cmdOut))
	return sha, nil
}

func GetGitTag(prefix string) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	} else {
		return GetGitRefFromDir(pwd)
	}
}

func GetGitTagFromDir(working_dir string, prefix string) (string, error) {
	var (
		cmdOut []byte
		err    error
	)

	cmdName := "git"
	cmdArgs := []string{"describe", "--match", prefix, "--abbrev=0", "--always"}
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = working_dir

	if cmdOut, err = cmd.Output(); err != nil {
		return "", errors.New("There was an error running git rev-parse command")
	}
	sha := strings.TrimSpace(string(cmdOut))
	return sha, nil
}
