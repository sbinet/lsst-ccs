// fcs-mgr manages a ccs+fcs-subsystem+localdb installation
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	svnRoot = "svn+ssh://svn.lsstcorp.org/camera/CameraControl"
)

var (
	repos = []string{
		"org-lsst-ccs-subsystem-fcs",
		"org-lsst-ccs-localdb",
	}
)

func main() {
	flag.Parse()
	if flag.NArg() <= 0 {
		flag.Usage()
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	err := dispatch(cmd, flag.Args()[1:])
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
}

func dispatch(cmd string, args []string) error {
	switch cmd {
	case "init":
		return cmdInit(args)
	case "build":
		return cmdBuild(args)
	case "localdb":
		return cmdLocalDB(args)
	case "update":
		return cmdUpdate(args)
	case "dist":
		return cmdDist(args)
	case "run":
		return cmdRun(args)
	default:
		return fmt.Errorf("unknown command %q\n", cmd)
	}

	panic("unreachable")
}
