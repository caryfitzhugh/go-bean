package main

import (
	"errors"
	//"fmt"
	//	"github.com/hashicorp/go-version"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const tag_prefix = "go-bean"

func GitAllCheckedIn() bool {
	pwd, err := os.Getwd()
	if err != nil {
		return false
	} else {
		return GitAllCheckedInInDir(pwd)
	}
}

func GitAllCheckedInInDir(working_dir string) bool {
	cmdName := "git"
	cmdArgs := []string{"diff", "--exit-code"}
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = working_dir
	if err := cmd.Run(); err != nil {
		return false
	}

	cmdName = "git"
	cmdArgs = []string{"diff", "--cached", "--exit-code"}
	cmd = exec.Command(cmdName, cmdArgs...)
	cmd.Dir = working_dir
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

func GitTagCurrent(ver string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	} else {
		return GitTagCurrentInDir(ver, pwd)
	}
}
func GitTagCurrentInDir(ver string, working_dir string) error {
	cmdName := "git"
	cmdArgs := []string{"tag", "-f", tag_prefix + "-" + ver}
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

func GetCurVer() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	} else {
		filename := pwd + "/VERSION"

		if _, err := os.Stat(filename); os.IsNotExist(err) {
			// path/to/whatever does not exist
			// So write and return "v0.0.0"
			_ = SaveCurVer("0.0.0")
		}

		ver, verr := ioutil.ReadFile(pwd + "/VERSION")
		if verr != nil {
			return "", verr
		} else {
			return string(ver), nil
		}
	}
}

func SaveCurVer(ver string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	} else {
		err = ioutil.WriteFile(pwd+"/VERSION", []byte(ver), 0644)
		if err != nil {
			return err
		} else {
			return nil
		}
	}
}
