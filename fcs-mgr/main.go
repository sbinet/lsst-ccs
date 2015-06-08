// fcs-mgr manages a ccs+fcs-subsystem+localdb installation
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
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
	case "update":
		return cmdUpdate(args)
	default:
		return fmt.Errorf("unknown command %q\n", cmd)
	}

	panic("unreachable")
}

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
			f, err := os.Create("log-" + repo + ".txt")
			if err != nil {
				errc <- err
				return
			}
			defer f.Close()
			rdir := filepath.Join(dir, repo)
			_, err = os.Stat(rdir)
			if err != nil {
				initRepo(rdir, errc)
			}

			updateRepo(rdir, errc)
			buildRepo(rdir, errc)
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

func cmdBuild(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("invalid number of arguments. got %d. want 0")
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

		go buildRepo(rdir, errc)
	}

	for range repos {
		err = <-errc
		if err != nil {
			return err
		}
	}

	return err
}

func cmdUpdate(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("invalid number of arguments. got %d. want 0")
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

		go updateRepo(rdir, errc)
	}

	for range repos {
		err = <-errc
		if err != nil {
			return err
		}
	}

	return err
}

func initRepo(rdir string, errc chan error) {
	repo := filepath.Base(rdir)
	log.Printf("init repo [%s]...\n", repo)

	f, err := os.Create("log-" + repo + ".txt")
	if err != nil {
		errc <- err
		return
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
	// cmd.Stdin = os.Stdin
	cmd.Stdout = f
	cmd.Stderr = f

	err = cmd.Run()
	if err != nil {
		errc <- err
		return
	}
	errc <- nil
}

func updateRepo(rdir string, errc chan error) {
	repo := filepath.Base(rdir)
	log.Printf("updating repo [%s]...\n", repo)

	f, err := os.Create("log-" + repo + ".txt")
	if err != nil {
		errc <- err
		return
	}
	defer f.Close()

	cmd := exec.Command("git", "svn", "fetch")
	cmd.Dir = rdir
	// cmd.Stdin = os.Stdin
	cmd.Stdout = f
	cmd.Stderr = f

	err = cmd.Run()
	if err != nil {
		errc <- err
		return
	}

	errc <- nil
}

func buildRepo(rdir string, errc chan error) {
	repo := filepath.Base(rdir)
	log.Printf("building repo [%s]...\n", repo)

	f, err := os.Create("log-" + repo + ".txt")
	if err != nil {
		errc <- err
		return
	}
	defer f.Close()

	cmd := exec.Command("fcs-boot", "-lsst="+rdir, "mvn", "clean", "install")
	cmd.Dir = rdir
	cmd.Stdin = os.Stdin
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
		errc <- err
		return
	}
	log.Printf("building repo [%s]... [ok] (time=%v)\n", repo, delta)
	errc <- nil
}
