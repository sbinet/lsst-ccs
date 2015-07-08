// fcs-mgr manages a ccs+fcs-subsystem+localdb installation
package main

import (
	"log"
	"os"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

const (
	svnRoot = "svn+ssh://svn.lsstcorp.org/camera/CameraControl"
)

var (
	repos = []string{
		"org-lsst-ccs-subsystem-fcs",
		"org-lsst-ccs-localdb",
	}

	app *commander.Command
)

func init() {
	app = &commander.Command{
		UsageLine: "fcs-mgr",
		Subcommands: []*commander.Command{
			fcsMakeCmdInit(),
			fcsMakeCmdBuild(),
			fcsMakeCmdLocalDB(),
			fcsMakeCmdUpdate(),
			fcsMakeCmdDist(),
			fcsMakeCmdDeploy(),
			fcsMakeCmdRun(),
		},
		Flag: *flag.NewFlagSet("fcs-mgr", flag.ExitOnError),
	}
}

func main() {
	err := app.Flag.Parse(os.Args[1:])
	if err != nil {
		log.Printf("error parsing flags: %v\n", err)
		os.Exit(1)
	}

	args := app.Flag.Args()
	err = app.Dispatch(args)
	if err != nil {
		log.Printf("error dispatching command: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
