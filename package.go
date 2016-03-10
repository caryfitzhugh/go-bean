package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

func TaggedName(host string, app_n string, ver string) string {
	return host + "/" + app_n + ":" + ver
}

func BuildBinary(program_name string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	} else {
		return BuildBinaryInDir(program_name, pwd)
	}
}

func BuildBinaryInDir(program_name string, working_dir string) error {
	cmdName := "go"
	cmdArgs := []string{"build", "-a", "-installsuffix", "cgo-linux-go-bean", "-o", program_name}

	env := os.Environ()
	env = append(env, "CGO_ENABLED=0")
	env = append(env, "GOOS=linux")

	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Env = env
	cmd.Dir = working_dir

	if err := cmd.Run(); err != nil {
		fmt.Printf("%v", err)
		return errors.New("There was an error running go build")
	}

	return nil
}

func BuildDockerImage(program_name string, program_port string, version_tag string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	} else {
		return BuildDockerImageInDir(program_name, program_port, version_tag, pwd)
	}
}

func BuildDockerImageInDir(program_name string, program_port string, version_tag string, working_dir string) error {
	file, tempfile_err := ioutil.TempFile(working_dir, "dockerfile")
	defer os.Remove(file.Name())

	if tempfile_err != nil {
		return tempfile_err
	}

	// Write out the tempfile
	dockerfile_contents := []byte("FROM scratch\nADD . /\nCMD [\"/" + program_name + "\"]\nEXPOSE " + program_port + "\n")
	tempfile_err = ioutil.WriteFile(file.Name(), dockerfile_contents, 0644)
	if tempfile_err != nil {
		return tempfile_err
	}

	cmdName := "sudo"
	cmdArgs := []string{"docker", "build", "-t", program_name + ":" + version_tag, "-f", file.Name(), "."}
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = working_dir

	if err := cmd.Run(); err != nil {
		fmt.Printf("%v", err)
		return errors.New("There was an error running docker build")
	}
	return nil
}

// Call ECR, get login information and execute
func LoginToECR() error {
	var (
		cmdOut []byte
		err    error
	)

	cmdName := "aws"
	cmdArgs := []string{"ecr", "get-login", "--region", "us-east-1"}
	cmd := exec.Command(cmdName, cmdArgs...)

	if cmdOut, err = cmd.Output(); err != nil {
		return errors.New("There was an error logging into AWS ECR. Do you have the aws CLI installed?")
	}

	cmdName = "sudo"
	cmdArgs = strings.Split(strings.TrimSpace(string(cmdOut)), " ")
	cmd = exec.Command(cmdName, cmdArgs...)

	if err := cmd.Run(); err != nil {
		fmt.Printf("%v", err)
		return errors.New("There was an error logging in the AWS docker registry")
	}
	return nil
}

// This will package the go repository at the working dir , into the given docker template
// And push the image to the host.
// It returns the string which is the path for the docker image
// i.e. 234234234234.ecr.amazon.com/program_name:$VERSION_TAG
func PushToRepository(program_name string, version_tag string, docker_host string) error {
	cmdName := "sudo"
	cmdArgs := []string{"docker", "tag", "-f", program_name + ":" + version_tag,
		TaggedName(docker_host, program_name, version_tag)}
	cmd := exec.Command(cmdName, cmdArgs...)

	if err := cmd.Run(); err != nil {
		fmt.Printf("%v", err)
		return errors.New("There was an error tagging the docker image")
	}

	cmdName = "sudo"
	cmdArgs = []string{"docker", "push", TaggedName(docker_host, program_name, version_tag)}
	cmd = exec.Command(cmdName, cmdArgs...)

	if err := cmd.Run(); err != nil {
		fmt.Printf("%v", err)
		return errors.New("There was an error pushing the docker image")
	}
	return nil
}

func GetAWSJsonFile(docker_tagged_name string, program_port string) []byte {
	return []byte("{\"AWSEBDockerrunVersion\": \"1\",\"Image\": { \"Name\": \"" +
		docker_tagged_name + "\", \"Update\": \"true\" }, \"Ports\":[{\"ContainerPort\":\"" + program_port + "\"}]}")
}

// This will take a docker image path, and inject it into a dockerrun.aws.json file.
// It uploads that file to S3, and then creates an application version for the application
// If all is well, you can use that to deploy on EB.
func CreateApplicationVersion(application_name string, version_tag string, program_name string,
	docker_host string,
	program_port string,
	s3_bucket string) (string, error) {

	// We want to generate an AWS.json.template file
	json_bytes := GetAWSJsonFile(TaggedName(docker_host, program_name, version_tag), program_port)
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	f, err := w.Create("Dockerrun.aws.json")
	if err != nil {
		return "", errors.New("There was an error creating the dockerrun.aws.json file")
	}
	_, err = f.Write(json_bytes)
	if err != nil {
		return "", errors.New("There was an error writing the dockerrun.aws.json file")
	}

	err = w.Close()
	if err != nil {
		return "", errors.New("There was an error closing the dockerrun.aws.json zip file")
	}
	zip_filename := application_name + "_" + version_tag + ".zip"
	zip_path := os.TempDir() + "/" + zip_filename
	ioutil.WriteFile(zip_path, buf.Bytes(), 0644)
	defer os.Remove(zip_path)

	// Now upload that file with the AWS CLI to S3
	cmd := exec.Command("aws", "s3", "cp", zip_path, "s3://"+s3_bucket)
	if err := cmd.Run(); err != nil {
		fmt.Printf("%v", err)
		return "", errors.New("There was an error uploading to S3")
	}

	// Now Create a new application version on AWS.
	cmd = exec.Command("aws", "elasticbeanstalk", "create-application-version",
		"--application-name", application_name,
		"--version-label", version_tag,
		"--source-bundle", "S3Bucket="+s3_bucket+",S3Key="+zip_filename)
	if err := cmd.Run(); err != nil {
		fmt.Printf("%v", err)
		return "", errors.New("There was an error uploading to S3")
	}
	return version_tag, nil
}

func UpdateEBEnvironment(env_name string, app_ver string) error {
	cmd := exec.Command("aws", "elasticbeanstalk", "update-environment",
		"--environment-name", env_name,
		"--version-label", app_ver)
	if err := cmd.Run(); err != nil {
		fmt.Printf("%v", err)
		return errors.New("There was an error updating the environment.  Does it exist?")
	}
	return nil
}

type DescribedEnvironment struct {
	Environments []EnvState
}
type EnvState struct {
	EnvironmentName string
	VersionLabel    string
	Status          string
}

func CurrentEnvState(env_name string) (EnvState, error) {
	var state EnvState
	cmdName := "aws"
	cmdArgs := []string{"elasticbeanstalk", "describe-environments", "--environment-names", env_name}

	cmd := exec.Command(cmdName, cmdArgs...)

	if cmdOut, err := cmd.Output(); err != nil {
		return state, errors.New("There was an error getting the AWS environment data")
	} else {
		var states DescribedEnvironment
		marshal_err := json.Unmarshal(cmdOut, &states)
		if marshal_err != nil {
			return state, marshal_err
		} else {
			return states.Environments[0], nil
		}
	}
}
func WaitForEBToBeReady(env_name string, version_label string) error {
	updated := false

	for updated {
		state, err := CurrentEnvState(env_name)
		if err != nil {
			return err
		}
		updated = state.Status == "Ready" &&
			state.VersionLabel == version_label
		print(".")
		time.Sleep(500 * time.Millisecond)
	}
	return nil
}
