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
	lsst = flag.String("lsst", "", "path to LSST FCS code tree (default=$PWD)")
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
		"run", "-it",
		"-p=50000:50000",
		"--user=" + usr.Uid + ":" + usr.Gid,
		"--net=host",
		"-v", *lsst + ":/opt/lsst",
	}

	if _, err := os.Stat(sshdir); err == nil {
		subcmd = append(subcmd, "-v", sshdir+":/home/lsst/.ssh")
	}
	if _, err := os.Stat(gopath); err == nil && gopath != "" {
		subcmd = append(subcmd, "-v", gopath+":/go")
	}

	subcmd = append(subcmd, "lsst-ccs/fcs")

	switch flag.NArg() {
	case 0:
		subcmd = append(subcmd, "/bin/bash")
	default:
		subcmd = append(subcmd, flag.Args()...)
	}

	cmd := exec.Command("docker", subcmd...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Fatalf("error running docker: %v\n", err)
	}
}
