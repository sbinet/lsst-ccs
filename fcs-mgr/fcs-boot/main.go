// fcs-boot runs the lsst-ccs/fcs docker image with a few options
package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

var (
	lsst     = flag.String("lsst", "", "path to LSST FCS code tree (default=$PWD)")
	tty      = flag.Bool("tty", true, "require a TTY")
	mysql    = flag.Bool("mysql", false, "connect to ccs-mysql container")
	detach   = flag.Bool("detach", false, "run container in background (daemonize)")
	name     = flag.String("name", "", "container name")
	memlimit = flag.String("memory", "", "memory limit of container")
)

func main() {
	flag.Parse()
	if *lsst == "" || *lsst == "." {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("could not retrieve $PWD: %v\n", err)
		}
		*lsst = wd
	}

	sshdir := filepath.Join(os.Getenv("HOME"), ".ssh")
	gopath := strings.Split(os.Getenv("GOPATH"), ":")[0]

	usr, err := user.Current()
	if err != nil {
		log.Fatalf("could not retrieve current user infos: %v\n", err)
	}

	subcmd := []string{
		"run", "-t",
		"-p=50000:50000", // for CANopen
		"-p=45566:45566", // for JBoss/JGroups
		"--user=" + usr.Uid + ":" + usr.Gid,
		"--net=host",
		"-v", *lsst + ":/opt/lsst",
		"-e", "DISPLAY=" + os.Getenv("DISPLAY"),
		"-v", "/tmp/.X11-unix:/tmp/.X11-unix",
	}

	if *mysql {
		subcmd = append(subcmd, "--volumes-from=ccs-mysql")
	}

	if *detach {
		subcmd = append(subcmd, "--detach")
	} else {
		subcmd = append(subcmd, "--rm")
	}

	if *name != "" {
		subcmd = append(subcmd, "--name", *name)
	}

	if *memlimit != "" {
		subcmd = append(subcmd, "--memory="+*memlimit)
	}

	if _, err := os.Stat(sshdir); err == nil {
		subcmd = append(subcmd, "-v", sshdir+":/home/lsst/.ssh")
	}
	if _, err := os.Stat(gopath); err == nil && gopath != "" {
		subcmd = append(subcmd, "-v", gopath+":/go")
	}

	if *tty {
		subcmd = append(subcmd, "-i")
	}

	subcmd = append(subcmd, "lsst-ccs/fcs")

	switch flag.NArg() {
	case 0:
		subcmd = append(subcmd, "/bin/bash")
	default:
		subcmd = append(subcmd, flag.Args()...)
	}

	enableX11()
	defer disableX11()

	cmd := exec.Command("docker", subcmd...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		disableX11()
		log.Fatalf("error running docker: %v\n", err)
	}
}

func enableX11() {
	err := runCmd("xhost", "+")
	if err != nil {
		log.Fatalf("error enabling X11-xhost: %v\n", err)
	}
}

func disableX11() {
	err := runCmd("xhost", "-")
	if err != nil {
		log.Fatalf("error disable X11-xhost: %v\n", err)
	}
}

func runCmd(cmd string, args ...string) error {
	exe := exec.Command(cmd, args...)
	exe.Stdin = os.Stdin
	exe.Stdout = os.Stdout
	exe.Stderr = os.Stderr
	return exe.Run()
}
