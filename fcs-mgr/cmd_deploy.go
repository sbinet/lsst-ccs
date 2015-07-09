package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func fcsMakeCmdDeploy() *commander.Command {
	cmd := &commander.Command{
		Run:       cmdDeploy,
		UsageLine: "deploy",
		Short:     "deploy sources and binaries to PC-104",
		Long: `
deploy deploys sources and binaries (FCS+c-wrapper) to the embedded PC-104.

ex:
 $ fcs-mgr deploy
`,
		Flag: *flag.NewFlagSet("fcs-mgr-deploy", flag.ExitOnError),
	}
	cmd.Flag.String("addr", "clrlsstemb01.in2p3.fr", "address of PC-104 where to deploy")
	cmd.Flag.String("user", "root", "PC-104 user")
	cmd.Flag.String("dir", "/opt/lsst", "directory where to deploy")
	return cmd
}

func cmdDeploy(cmdr *commander.Command, args []string) error {
	var err error
	if len(args) > 0 {
		return fmt.Errorf(
			"invalid number of arguments. got %d. want 0",
			len(args),
		)
	}

	flog, err := os.Create("log-deploy.txt")
	if err != nil {
		log.Fatalf("error creating logfile: %v\n", err)
	}
	defer flog.Close()

	run := func(cmd *exec.Cmd) error {
		_, err := fmt.Fprintf(flog, "\n### %s %v\n", cmd.Path, cmd.Args)
		if err != nil {
			return err
		}
		cmd.Stdin = os.Stdin
		cmd.Stdout = flog
		cmd.Stderr = flog
		return cmd.Run()
	}

	addr := cmdr.Flag.Lookup("addr").Value.Get().(string)
	odir := cmdr.Flag.Lookup("dir").Value.Get().(string)
	user := cmdr.Flag.Lookup("user").Value.Get().(string)

	uri := user + "@" + addr

	log.Printf("remote: %v:%v\n", uri, odir)

	cmd := exec.Command("ssh", uri, "mkdir", "-p", odir)
	err = run(cmd)
	if err != nil {
		log.Fatalf("error creating directory [%s] on remote: %v\n", odir, err)
	}

	log.Printf("creating cwrapper archive...\n")
	cmd = exec.Command("git", "archive", "-o", "../cwrapper.tar.gz", "HEAD")
	cmd.Dir = "cwrapper-git"
	err = run(cmd)
	if err != nil {
		log.Fatalf("error creating cwrapper archive: %v\n", err)
	}

	xfer := func(dir string) error {
		_, err := os.Stat(dir)
		if err != nil {
			log.Fatalf("could not stat [%s]: %v\n", dir, err)
		}

		cmd := exec.Command(
			"scp",
			"-C",
			"-r", dir, uri+":"+odir,
		)
		err = run(cmd)
		if err != nil {
			log.Printf("error transferring [%s]: %v\n", dir, err)
		}
		return err
	}

	{
		dir := "cwrapper.tar.gz"
		log.Printf("transferring [%s]...\n", dir)
		err = xfer(dir)
		if err != nil {
			return err
		}

		err = os.Remove(dir)
		if err != nil {
			log.Printf("error cleaning up [%s]: %v\n", dir, err)
			return err
		}
	}

	log.Printf("uncompressing cwrapper.tar.gz...\n")
	cmd = exec.Command("ssh", uri,
		"cd "+odir+"; rm -rf cwrapper; "+
			"mkdir cwrapper;"+
			"cd cwrapper && tar zxf ../cwrapper.tar.gz &&"+
			"cd .. && rm -rf cwrapper.tar.gz",
	)
	err = run(cmd)
	if err != nil {
		log.Fatalf("error uncompressing cwrapper: %v\n", err)
	}

	log.Printf("compiling cwrapper...\n")
	cmd = exec.Command("ssh", uri, "cd "+odir+"/cwrapper/src/cwrapper-lpc; make")
	start := time.Now()
	err = run(cmd)
	delta := time.Since(start)
	if err != nil {
		log.Fatalf("error compiling cwrapper: %v (%v)\n", err, delta)
	}
	log.Printf("compiling cwrapper... [done] (%v)\n", delta)
	if err == nil {
		flog.Close()
		os.Remove(flog.Name())
	}

	return err
}
