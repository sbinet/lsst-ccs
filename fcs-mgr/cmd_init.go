package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func cmdInit(args []string) error {
	var err error
	dir := "."
	switch len(args) {
	case 0:
		dir, err = os.Getwd()
		if err != nil {
			return err
		}
	case 1:
		dir, err = filepath.Abs(args[0])
		if err != nil {
			log.Printf("error expanding path [%s]: %v\n", args[0], err)
			return err
		}

	default:
		err = fmt.Errorf(
			"invalid number of arguments. init expects 0 or 1 (got %d)",
			len(args),
		)
	}
	log.Printf("init-dir=%q\n", dir)
	_, err = os.Stat(dir)
	if err == os.ErrNotExist {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			log.Printf(
				"error: could not create directory [%s] (err=%v)\n",
				dir,
				err,
			)
			return err
		}
	}

	errc := make(chan error)
	// for each sub-repo, import from svn (to git) if not already done.
	for _, repo := range repos {
		go func(repo string) {
			rdir := filepath.Join(dir, repo)
			_, err = os.Stat(rdir)
			if err != nil {
				err = initRepo(rdir)
				if err != nil {
					errc <- err
					return
				}
			}

			errc <- updateRepo(rdir)
		}(repo)
	}

	for range repos {
		err = <-errc
		if err != nil {
			return err
		}
	}

	return err
}

func initRepo(rdir string) error {
	repo := filepath.Base(rdir)
	log.Printf("init repo [%s]...\n", repo)

	f, err := os.Create("log-" + repo + ".txt")
	if err != nil {
		return err
	}
	defer f.Close()

	cmd := exec.Command(
		"git", "svn", "init",
		"--prefix=svn/", "--trunk=trunk",
		"--branches=branches",
		"--tags=tags",
		svnRoot+"/"+repo,
		repo,
	)
	//cmd.Stdin = os.Stdin
	cmd.Stdout = f
	cmd.Stderr = f

	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
