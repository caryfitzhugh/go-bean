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
	if DoesConfigExist() {
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
	fmt.Print("Application name on EB:  ")
	text, _ := reader.ReadString('\n')
	new_conf.AppName = strings.Trim(text, " \n")

	write_err := WriteConfig(new_conf)
	if write_err != nil {
		return errors.New("Failed to write config file")
	}
	return nil
}
