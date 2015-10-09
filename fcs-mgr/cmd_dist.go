package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

var (
	dbInfo = struct {
		User     string
		Pass     string
		RootPass string
	}{
		User:     "user",
		Pass:     "s3cr37",
		RootPass: "sup3r-s3cr37",
	}
)

func fcsMakeCmdDist() *commander.Command {
	cmd := &commander.Command{
		Run:       cmdDist,
		UsageLine: "dist",
		Short:     "build a binary distribution kit",
		Long: `
dist builds a binary distribution kit.

ex:
 $ fcs-mgr dist
`,
		Flag: *flag.NewFlagSet("fcs-mgr-dist", flag.ExitOnError),
	}
	return cmd
}

func cmdDist(cmdr *commander.Command, args []string) error {
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
	for i := range repos {
		repo := &repos[i]
		rdir := filepath.Join(dir, repo.Name)
		_, err = os.Stat(rdir)
		if err != nil {
			log.Printf("no such directory [%s] (err=%v)\n", rdir, err)
			return err
		}

		go func(rdir string, repo *Repo) {
			errc <- makeDistRepo(dist, rdir, repo)
		}(rdir, repo)
	}

	for range repos {
		err = <-errc
		if err != nil {
			return err
		}
	}

	extdir := filepath.Join(dist, "externalResources")
	err = os.MkdirAll(extdir, 0755)
	if err != nil {
		log.Printf("error creating [%s] directory: %v\n", extdir, err)
		return err
	}

	drvdir := filepath.Join(dist, "drivers")
	err = os.MkdirAll(drvdir, 0755)
	if err != nil {
		log.Printf("error creating [%s] directory: %v\n", drvdir, err)
		return err
	}

	// FIXME(sbinet) extract/infer correct name
	mysqlConnector := filepath.Join(
		"..", repos[1].Name+"-main-"+repos[1].Version,
		"share", "java",
		"mysql-connector-java-5.1.23.jar",
	)
	err = os.Symlink(
		mysqlConnector,
		filepath.Join(drvdir, "mysql-connector-java.jar"),
	)
	if err != nil {
		log.Printf("error creating mysql-connector symlink: %v\n",
			err,
		)
		return err
	}

	for _, v := range []struct {
		Name string
		Data []byte
	}{
		{
			Name: filepath.Join(extdir, "statusPersister.properties"),
			Data: []byte(fmt.Sprintf(
				`hibernate.connection.url=jdbc:mysql://localhost:3306/ccs
hibernate.connection.driver_class=com.mysql.jdbc.Driver
hibernate.dialect=org.hibernate.dialect.MySQLDialect
hibernate.connection.username=%s
hibernate.connection.password=%s
`,
				dbInfo.User,
				dbInfo.Pass,
			)),
		},
		{
			Name: filepath.Join(extdir, "ccsGlobal.properties"),
			Data: []byte(fmt.Sprintf(
				"org.lsst.ccs.localdb.additional.classpath.entry=%s\n",
				"/opt/lsst/DISTRIB/drivers/mysql-connector-java.jar",
			)),
		},
	} {
		err = ioutil.WriteFile(v.Name, v.Data, 0644)
		if err != nil {
			log.Fatalf("error writing file [%s]: %v\n", v.Name, err)
			return err
		}
	}

	return err
}

func makeDistRepo(dist, rdir string, repo *Repo) error {
	log.Printf("creating distribution for repo [%s]...\n", repo.Name)

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

	f, err := os.Create("log-" + repo.Name + ".txt")
	if err != nil {
		return err
	}
	defer f.Close()

	root := filepath.Join(rdir, "*/target/*-"+data.Version+"-dist.zip")
	matches, err := filepath.Glob(root)
	if err != nil {
		return err
	}

	repo.Version = data.Version

	for _, fname := range matches {
		//fmt.Printf(">>> %s\n", fname)
		err = unzip(dist, fname)
		if err != nil {
			log.Printf("error unzip-ing [%s] into [%s]: %v\n", fname, dist, err)
			return err
		}
	}
	log.Printf("creating distribution for repo [%s]... [done]\n", repo.Name)
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
