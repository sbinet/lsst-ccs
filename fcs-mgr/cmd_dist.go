package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func cmdDist(args []string) error {
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

	dist := filepath.Join(dir, "DISTRIB")
	os.RemoveAll(dist)
	err = os.MkdirAll(dist, 0755)
	if err != nil {
		log.Printf("error creating top-level DISTRIB directory: %v\n", err)
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
			errc <- makeDistRepo(dist, rdir)
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

func makeDistRepo(dist, rdir string) error {
	repo := filepath.Base(rdir)
	log.Printf("creating distribution for repo [%s]...\n", repo)

	pom, err := os.Open(filepath.Join(rdir, "pom.xml"))
	if err != nil {
		return err
	}
	defer pom.Close()

	type POM struct {
		XMLName xml.Name `xml:"project"`
		Name    string   `xml:"name"`
		Version string   `xml:"version"`
	}
	var data POM
	err = xml.NewDecoder(pom).Decode(&data)
	if err != nil {
		log.Printf("error decoding pom.xml: %v\n", err)
		return err
	}

	// log.Printf("pom-data: %#v\n", data)

	f, err := os.Create("log-" + repo + ".txt")
	if err != nil {
		return err
	}
	defer f.Close()

	root := filepath.Join(rdir, "*/target/*-"+data.Version+"-dist.zip")
	matches, err := filepath.Glob(root)
	if err != nil {
		return err
	}

	for _, fname := range matches {
		//fmt.Printf(">>> %s\n", fname)
		err = unzip(dist, fname)
		if err != nil {
			log.Printf("error unzip-ing [%s] into [%s]: %v\n", fname, dist, err)
			return err
		}
	}
	log.Printf("creating distribution for repo [%s]... [done]\n", repo)
	return err
}

func unzip(dest, src string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			log.Fatalf("error closing zip archive: %v\n", err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extract := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				log.Fatalf("error closing zip file: %v\n", err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					log.Fatalf("error closing zip output file: %v\n", err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extract(f)
		if err != nil {
			return err
		}
	}

	return nil
}
