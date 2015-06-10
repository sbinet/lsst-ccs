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
		subargs = append(subargs, "shell")
	} else {
		subargs = append(subargs, args...)
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
