package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

func Command_Init(force_command bool) error {
	// If there is an existing config file
	current_config, config_err := LoadConfig()
	if config_err != nil {
		if force_command {
			if EraseConfig() != nil {
				return errors.New("Could not delete config file")
			}
		} else {
			return errors.New("Config already exists.  To overwrite pass the --force flag")
		}
	}

	reader := bufio.NewReader(os.Stdin)
	var new_conf = GoBeanConfig{}

	// Now we can create our file. Ask what application name this is connected with
	fmt.Print("Application name on EB:  [" + current_config.AppName + "]")
	text, _ := reader.ReadString('\n')
	new_conf.AppName = strings.Trim(text, " \n")
	if new_conf.AppName == "" {
		new_conf.AppName = current_config.AppName
	}

	fmt.Print("Staging environment name on EB:  [" + current_config.EnvName + "]")
	text, _ = reader.ReadString('\n')
	new_conf.EnvName = strings.Trim(text, " \n")
	if new_conf.EnvName == "" {
		new_conf.EnvName = current_config.EnvName
	}

	fmt.Print("S3 Bucket to store images:  [" + current_config.S3Bucket + "]")
	text, _ = reader.ReadString('\n')
	new_conf.S3Bucket = strings.Trim(text, " \n")
	if new_conf.S3Bucket == "" {
		new_conf.S3Bucket = current_config.S3Bucket
	}

	fmt.Print("Docker Host (AWS uses ECR):  [" + current_config.DockerHost + "]")
	text, _ = reader.ReadString('\n')
	new_conf.DockerHost = strings.Trim(text, " \n")
	if new_conf.DockerHost == "" {
		new_conf.DockerHost = current_config.DockerHost
	}

	fmt.Print("Program Name: (what is the binary called?)  [" + current_config.ProgramName + "]")
	text, _ = reader.ReadString('\n')
	new_conf.ProgramName = strings.Trim(text, " \n")
	if new_conf.ProgramName == "" {
		new_conf.ProgramName = current_config.ProgramName
	}

	fmt.Print("Program Port: (what is the port exposed?)  [" + current_config.ProgramPort + "]")
	text, _ = reader.ReadString('\n')
	new_conf.ProgramPort = strings.Trim(text, " \n")
	if new_conf.ProgramPort == "" {
		new_conf.ProgramPort = current_config.ProgramPort
	}

	write_err := WriteConfig(new_conf)
	if write_err != nil {
		return errors.New("Failed to write config file")
	}
	return nil
}
