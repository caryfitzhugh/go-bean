// Package go-bean/config is
package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

const config_file_name = ".go-bean"

type GoBeanConfig struct {
	AppName string
}

func ConfigFilePath() string {
	pwd, err := os.Getwd()
	if err != nil {
		println(err)
		os.Exit(1)
	}
	return filepath.Join(pwd, config_file_name)
}

func DoesConfigExist() bool {
	_, err := LoadConfig()
	return err == nil
}

func EraseConfig() error {
	config_file_path := ConfigFilePath()
	err := os.Remove(config_file_path)
	return err
}

func WriteConfig(conf GoBeanConfig) error {
	config_file_path := ConfigFilePath()
	bytes, err := json.Marshal(conf)
	if err != nil {
		return err
	}
	write_err := ioutil.WriteFile(config_file_path, bytes, 0644)

	return write_err
}

func LoadConfig() (GoBeanConfig, error) {
	var conf GoBeanConfig
	config_file_path := ConfigFilePath()

	bytes, err := ioutil.ReadFile(config_file_path)
	if err != nil {
		return conf, errors.New("Config file not found at [" + config_file_path + "], you must run init")

	} else {

		marshal_err := json.Unmarshal(bytes, &conf)
		if marshal_err != nil {
			return conf, errors.New("Parsing config file [" + config_file_path + "] failed")
		}

		return conf, nil
	}
}

func Command_Config() error {
	conf, err := LoadConfig()
	if err != nil {
		return err
	} else {
		println("go-bean Configuration: ")
		println("  Application Name: " + conf.AppName)
		println("")
		return nil
	}
}
