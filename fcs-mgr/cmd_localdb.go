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

const (
	dkrMysql   = "ccs-mysql"
	dkrLocaldb = "ccs-localdb"
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

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// try to create the ccs-mysql container.
	// if it already exists, docker should tell us.
	cmd := exec.Command(
		"docker", "run", "--detach",
		"--env", "MYSQL_ROOT_PASSWORD="+dbInfo.RootPass,
		"--env", "MYSQL_USER="+dbInfo.User,
		"--env", "MYSQL_PASSWORD="+dbInfo.Pass,
		"--env", "MYSQL_DATABASE=ccs",
		"--name", dkrMysql,
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
	mysql, err := dockerContainer(dkrMysql)
	if err != nil {
		if mysql.Id == "N/A" {
			log.Printf("%s container is NOT RUNNING.\n", dkrMysql)
			log.Printf("please run 'fcs-mgr localdb create' first\n")
			return fmt.Errorf("%s container is NOT RUNNING", dkrMysql)
		}
		return err
	}

	if !mysql.State.Running {
		log.Printf("%s container is NOT RUNNING: %#v\n", dkrMysql, mysql)
		log.Printf("please run 'fcs-mgr localdb create' first\n")
		return fmt.Errorf("%s container is NOT RUNNING", dkrMysql)
	}

	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"fcs-boot",
		"-mysql", "-lsst="+dir, "-detach",
		"-name="+dkrLocaldb,
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

	run := func(cmd string, args ...string) error {
		exe := exec.Command(cmd, args...)
		exe.Stdin = os.Stdin
		exe.Stdout = os.Stdout
		exe.Stderr = os.Stderr
		return exe.Run()
	}

	for _, name := range []string{dkrLocaldb, dkrMysql} {
		container, err := dockerContainer(name)
		if err != nil {
			log.Printf("error retrieving status of container %s: %v\n", name,
				err)
			return err
		}

		err = run("docker", "stop", container.Id)
		if err != nil {
			log.Printf("could not stop %s container: %v\n", name, err)
			return err
		}

		err = run("docker", "rm", container.Id)
		if err != nil {
			log.Printf("could not remove %s container: %v\n", name, err)
			return err
		}
	}

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
