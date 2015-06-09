package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func cmdBuild(args []string) error {
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
		rdir := filepath.Join(dir, repo)
		_, err = os.Stat(rdir)
		if err != nil {
			log.Printf("no such directory [%s] (err=%v)\n", rdir, err)
			return err
		}

		go func(rdir string) {
			errc <- buildRepo(rdir)
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

func buildRepo(rdir string) error {
	repo := filepath.Base(rdir)
	log.Printf("building repo [%s]...\n", repo)

	f, err := os.Create("log-" + repo + ".txt")
	if err != nil {
		return err
	}
	defer f.Close()

	cmd := exec.Command("fcs-boot", "-tty=false", "-lsst="+rdir, "mvn", "clean", "install")
	//cmd.Dir = rdir
	//cmd.Stdin = f
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
