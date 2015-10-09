package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func fcsMakeCmdBuild() *commander.Command {
	cmd := &commander.Command{
		Run:       cmdBuild,
		UsageLine: "build",
		Short:     "build the code in a FCS/CCS workarea",
		Long: `
build builds the code in a FCS/CCS workarea.

ex:
 $ fcs-mgr build
`,
		Flag: *flag.NewFlagSet("fcs-mgr-build", flag.ExitOnError),
	}
	cmd.Flag.Bool("test", false, "run tests")
	return cmd
}

func cmdBuild(cmdr *commander.Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf(
			"invalid number of arguments. got %d. want 0",
			len(args),
		)
	}

	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	errc := make(chan error)
	for _, repo := range repos {
		rdir := filepath.Join(dir, repo.Name)
		_, err = os.Stat(rdir)
		if err != nil {
			log.Printf("no such directory [%s] (err=%v)\n", rdir, err)
			return err
		}

		go func(rdir string) {
			errc <- buildRepo(rdir, cmdr.Flag.Lookup("test").Value.Get().(bool))
		}(rdir)

	}

	for range repos {
		err = <-errc
		if err != nil {
			return err
		}
	}

	return err
}

func buildRepo(rdir string, tests bool) error {
	skip := !tests
	repo := filepath.Base(rdir)
	log.Printf("building repo [%s]...\n", repo)

	f, err := os.Create("log-" + repo + ".txt")
	if err != nil {
		return err
	}
	defer f.Close()

	cmd := exec.Command(
		"fcs-boot", "-name=ccs-build-"+repo, "-tty=false", "-lsst="+rdir,
		"mvn",
		fmt.Sprintf("-Dmaven.test.skip=%v", skip),
		"clean", "install",
	)
	cmd.Stdout = f
	cmd.Stderr = f

	start := time.Now()
	err = cmd.Run()
	delta := time.Since(start)
	if err != nil {
		log.Printf(
			"building repo [%s]... [err=%v] (time=%v)\n",
			repo,
			err,
			delta,
		)
		return err
	}
	log.Printf("building repo [%s]... [ok] (time=%v)\n", repo, delta)
	return nil
}
