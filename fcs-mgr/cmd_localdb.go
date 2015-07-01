package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

type DockerContainer struct {
	Id    string
	State struct {
		Running    bool
		Paused     bool
		Restarting bool
		OOMKilled  bool
		Dead       bool
		Pid        int
		ExitCode   int
		Error      string
		StartedAt  time.Time
		FinishedAt time.Time
	} `json:"State"`
}

func cmdLocalDB(args []string) error {
	// all the localdb subcommands will need to use docker somehow
	// make sure it is accessible
	_, err := exec.LookPath("docker")
	if err != nil {
		log.Printf("could not locate 'docker' command: %v\n", err)
		return err
	}

	switch args[0] {
	case "create":
		return cmdLocalDBCreate(args[1:])
	case "start":
		return cmdLocalDBStart(args[1:])
	case "stop":
		return cmdLocalDBStop(args[1:])
	default:
		return fmt.Errorf("unknown localdb command %q\n", args[0])
	}
	panic("unreachable")
}

func cmdLocalDBCreate(args []string) error {
	var err error

	// is the container already running? created?
	docker, err := dockerContainer("ccs-mysql")
	if err == nil {

		// container 'ccs-mysql' exists.
		// restart it if needed or do nothing (if already running)

		status := docker.State
		switch {
		case status.Running:
			log.Printf("localdb container already running\n")
			return nil

		case status.Restarting:
			log.Printf("localdb container is restarting... (retry later)\n")
			return nil

		case status.Paused:
			log.Printf("localdb container paused. re-starting\n")
			cmd := exec.Command("docker", "restart", docker.Id)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			return err

		case status.OOMKilled:
			log.Printf("localdb container killed (OOM)\n")
			return fmt.Errorf("localdb container killed (OOM)")

		case status.Dead:
			log.Printf("localdb container is dead\n")
			return fmt.Errorf("localdb container is dead")

		default:
			log.Printf("localdb container in UNKNOWN state:\n%v\n", docker)
			return fmt.Errorf("localdb container in UNKNOWN state")
		}

		log.Printf(">>> data=%#v\n", status)
	}
	if err != nil && docker.Id != "N/A" {
		return err
	}

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"docker", "run", "--detach",
		"--env", "MYSQL_ROOT_PASSWORD="+dbInfo.RootPass,
		"--env", "MYSQL_USER="+dbInfo.User,
		"--env", "MYSQL_PASSWORD="+dbInfo.Pass,
		"--env", "MYSQL_DATABASE=ccs",
		"--name", "ccs-mysql",
		"--publish", "3306:3306",
		"--volume", pwd+"/mysql:/var/lib/mysql",
		"lsst-ccs/mysql",
	)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	return err
}

func cmdLocalDBStart(args []string) error {
	var err error
	// make sure 'ccs-mysql' is running
	docker, err := dockerContainer("ccs-mysql")
	if err != nil {
		return err
	}

	if !docker.State.Running {
		log.Printf("localdb container is NOT RUNNING: %#v\n", docker)
		log.Printf("please run 'fcs-mgr localdb create' first\n")
		return fmt.Errorf("localdb container is NOT RUNNING")
	}

	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"fcs-boot",
		"-mysql", "-lsst="+dir, "-detach",
		"-name=ccs-localdb",
		"fcs-run",
		"start-localdb",
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return err
}

func cmdLocalDBStop(args []string) error {
	var err error
	return err
}

func dockerContainer(name string) (DockerContainer, error) {
	// is the container already running? created?
	cmd := exec.Command("docker", "inspect", name)

	out := new(bytes.Buffer)
	cmd.Stdin = os.Stdin
	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		// container does not exist
		return DockerContainer{Id: "N/A"}, err
	}

	data := []DockerContainer{}
	err = json.NewDecoder(out).Decode(&data)
	if err != nil {
		return DockerContainer{}, err
	}
	if len(data) != 1 {
		return DockerContainer{}, fmt.Errorf("invalid docker inspect output: %#v\n", data)
	}

	return data[0], nil
}
