package main

import (
	"os"
	"os/exec"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func fcsMakeCmdRun() *commander.Command {
	cmd := &commander.Command{
		Run:       cmdRun,
		UsageLine: "run [cmd [args...]]",
		Short:     "run a command inside a CCS container",
		Long: `
run runs a command inside a container with the CCS framework mounted under
/opt/lsst and via the fcs-boot command.

ex:
 $ fcs-mgr run lpc
 $ fcs-mgr run shell
`,
		Flag: *flag.NewFlagSet("fcs-mgr-", flag.ExitOnError),
	}
	return cmd
}

func cmdRun(cmdr *commander.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	subargs := []string{"-mysql", "-lsst=" + dir, "fcs-run"}
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
