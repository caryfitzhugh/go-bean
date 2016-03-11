package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func RunCommand(cmd *exec.Cmd, behavior string) (string, error) {
	var cmd_output bytes.Buffer
	var cmd_stderr bytes.Buffer
	cmd.Stdout = &cmd_output
	cmd.Stderr = &cmd_stderr

	if err := cmd.Run(); err != nil {
		return "", errors.New("There was an error " + behavior + ":\n\n" +
			fmt.Sprint("%v", cmd) + "\n" +
			fmt.Sprint(err) + ": " + cmd_stderr.String())
	}
	return cmd_output.String(), nil
}

func check_err(err error, message string) {
	if err != nil {
		println("ERROR")
		println(message)
		println("")
		os.Exit(1)
	}
}
