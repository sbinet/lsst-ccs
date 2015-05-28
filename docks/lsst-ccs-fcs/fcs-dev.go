// fcs-dev runs the lsst-ccs/fcs docker image with a few options
package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
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

	subcmd := []string{
		"run", "-it",
		"-v", *lsst + ":/opt/lsst",
		"-v", sshdir + ":/home/lsst/.ssh",
		"-v", gopath + ":/go",
		"-P",
		"-p=50000:50000",
		"--net=host",
		"lsst-ccs/fcs",
	}
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

	err := cmd.Run()
	if err != nil {
		log.Fatalf("error running docker: %v\n", err)
	}
}
