package main

import (
	"os"
	"os/exec"
)

func cmdRun(args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	subargs := []string{"-lsst=" + dir, "fcs-run"}
	if len(args) <= 0 {
		subargs = append(subargs, "bash")
	} else {
		subargs = append(subargs, args...)
	}

	// make sure the distrib is up-to-date
	err = cmdDist(nil)
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"fcs-boot",
		subargs...,
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	return err
}
