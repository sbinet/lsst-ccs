package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func cmdUpdate(args []string) error {
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
			errc <- updateRepo(rdir)
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

func updateRepo(rdir string) error {
	repo := filepath.Base(rdir)
	log.Printf("updating repo [%s]...\n", repo)

	f, err := os.Create("log-" + repo + ".txt")
	if err != nil {
		return err
	}
	defer f.Close()

	cmd := exec.Command("git", "svn", "fetch")
	cmd.Dir = rdir
	//cmd.Stdin = os.Stdin
	cmd.Stdout = f
	cmd.Stderr = f

	err = cmd.Run()
	if err != nil {
		log.Printf("updating repo [%s]... [err=%v]\n", repo, err)
		return err
	}

	return nil
}
