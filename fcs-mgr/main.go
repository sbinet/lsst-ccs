package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/xml"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

var (
	haddr = flag.String("addr", "", "<ip>[:<port>] PC-104 will listen to")

	host = ""
	port = 50000

	exitFuncs  = []func(){}
	wdir       = "."
	fcsVersion = "dev-SNAPSHOT"
)

func main() {

	var err error

	defer func() {
		os.Stderr.Sync()
		os.Stdout.Sync()
		runExitFuncs()
	}()

	flag.Parse()
	switch *haddr {
	case "":
		host = getHostIP()
	default:
		slice := strings.Split(*haddr, ":")
		switch len(slice) {
		case 2:
			host = slice[0]
			port, err = strconv.Atoi(slice[1])
			if err != nil {
				fatalf("invalid port number: %s. (err=%v)\n", slice[1], err)
			}
		case 1:
			host = slice[0]
		default:
			fatalf("invalid addr argument. got: %q\n", *haddr)
		}

		if host == "" {
			host = getHostIP()
		}
	}

	go func() {
		sigch := make(chan os.Signal)
		signal.Notify(sigch, os.Interrupt, os.Kill)
		for {
			select {
			case <-sigch:
				os.Stderr.Sync()
				os.Stdout.Sync()
				runExitFuncs()
				os.Stdout.Sync()
				os.Stderr.Sync()
			}
		}
	}()

	run()
}

func run() {
	var err error
	wdir, err = os.Getwd()
	if err != nil {
		fatalf("could not retrieve current work directory: %v\n", wdir)
	}

	errc := make(chan error)

	go startCWrapper(errc)
	initProject()
	makeDistrib()
	setupEnv()
	go dispatch(errc)

	select {
	case err := <-errc:
		if err != nil {
			fatalf("error: %v\n", err)
		}
	}
}

func atexit(f func()) {
	exitFuncs = append(exitFuncs, f)
}

func runExitFuncs() {
	log.Printf("running atexit-funcs...\n")
	for _, f := range exitFuncs {
		f()
	}
	log.Printf("running atexit-funcs... [done]\n")
}

func fatalf(format string, args ...interface{}) {
	runExitFuncs()
	log.Fatalf(format, args...)
}

func runCmd(cmd string, args ...string) error {
	exe := exec.Command(cmd, args...)
	exe.Stdin = os.Stdin
	exe.Stdout = os.Stdout
	exe.Stderr = os.Stderr
	atexit(func() {
		killProc(exe)
	})
	return exe.Run()
}

func killProc(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	state := cmd.ProcessState
	if state != nil && state.Exited() {
		return
	}
	p := cmd.Process
	pgid, err := syscall.Getpgid(p.Pid)
	if err != nil {
		log.Printf("could not get process group-id of [%v]: %v\n",
			cmd, err,
		)
		return
	}

	err = syscall.Kill(-pgid, syscall.SIGKILL)
	if err != nil {
		log.Printf("could not kill process [%v]: %v\n",
			cmd, err,
		)
	}
}

func getHostIP() string {
	host, err := os.Hostname()
	if err != nil {
		fatalf("could not retrieve hostname: %v\n", err)
	}

	addrs, err := net.LookupIP(host)
	if err != nil {
		fatalf("could not lookup hostname IP: %v\n", err)
	}

	for _, addr := range addrs {
		ipv4 := addr.To4()
		if ipv4 == nil {
			continue
		}
		return ipv4.String()
	}

	fatalf("could not infer host IP")
	return ""
}

func startCWrapper(errc chan error) {
	log.Printf("Starting c-wrapper on PC-104... (listen for %s:%d)\n", host, port)

	cmd := exec.Command(
		"ssh",
		"-X",
		"root@clrlsstemb01.in2p3.fr",
		"startCWrapper --host="+host,
	)
	cmd.Env = append(cmd.Env, "TERM=vt100")
	//cmd.Stdin = os.Stdin
	//cmd.Stdout = os.Stderr
	//cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	atexit(func() {
		killProc(cmd)
	})

	log.Printf("c-wrapper command: %v\n", cmd.Args)

	err := cmd.Run()
	if err != nil {
		log.Printf("c-wrapper= %v\n", err)
		errc <- err
	}
}

func initProject() {
	fname := filepath.Join(wdir, "pom.xml")
	pom, err := os.Open(fname)
	if err != nil {
		fatalf("could not open [%s]: %v\n", fname, err)
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
		fatalf("error decoding pom.xml: %v\n", err)
	}

	log.Printf("pom-data: %#v\n", data)
	fcsVersion = data.Version
}

func makeDistrib() {
	dist := filepath.Join(wdir, "DISTRIB")
	log.Printf("creating DISTRIB [%s]...\n", dist)
	os.RemoveAll(dist)
	err := os.MkdirAll(dist, 0755)
	if err != nil {
		fatalf("error creating [%s]: %v\n", dist, err)
	}

	for _, d := range []string{"main", "gui"} {
		dir := filepath.Join(wdir, d)
		_, err = os.Stat(dir)
		if err != nil {
			fatalf("could not stat [%s]: %v\n", dir, err)
		}

		srcs, err := filepath.Glob(
			filepath.Join(dir, "target", "org-lsst-ccs-subsystem-fcs-"+d+"-*-dist.zip"),
		)
		if err != nil {
			fatalf("could not stat %s-*-dist.zip: %v\n", d, err)
		}
		switch len(srcs) {
		case 0:
			fatalf("no %s-*-dist.zip!", d)
		case 1:
			// ok
		default:
			fatalf("too many %s-*-dist.zip files (%d): %v\n", len(srcs), srcs)
		}

		err = unzip(dist, srcs[0])
		if err != nil {
			fatalf("could not unzip [%s] into [%s]: %v\n", srcs[0], dist, err)
		}
	}
	log.Printf("creating DISTRIB [%s]... [done]\n", dist)
}

func setupEnv() {
	for _, v := range [][2]string{
		{"FCS_VERSION", fcsVersion},
		{"LOCALDB_VERSION", "1.3.0-SNAPSHOT"},
		// logging
		{"P0", "org.lsst.ccs.utilities.logging.ConsoleHandlerN.level=ALL"},
		{"P1", "org.lsst.ccs.utilities.logging.FileHandlerN.formatter=org.lsst.ccs.utilities.logging.TextFormatter"},
		{"P2", ".level=WARNING"},
		{"P3", "org.lsst.ccs.level=ALL"},
		{"P4", "org.lsst.ccs.bus.level=INFO"},
		{"P5", "org.lsst.ccs.subsystems.fcs.level=ALL"},

		{"TEST_ENV_ROOT", filepath.Join(wdir, "DISTRIB")},
		{"TEST_ENV_BIN", filepath.Join("${TEST_ENV_ROOT}", "bin")},
		{"TEST_ENV_RESOURCES", filepath.Join("${TEST_ENV_ROOT}", "externalResources")},
		{"TEST_ENV_DIRVERS", filepath.Join("${TEST_ENV_ROOT}", "drivers")},

		// needed by CCSbootstrap.sh
		{"CCS_RESOURCE_PATH", "${TEST_ENV_RESOURCES}"},

		// make sure we use IPv4 in JAS
		{"JASJVM_OPTS", "-Djava.net.preferIPV4Stack=true"},

		{"DESCRIPTION_FILE", "testbenchLPC.groo"},
		{"CONFIGURATION_FILE", "testbenchLPC_XXXXX_.properties"},
		{"WORKDIR", filepath.Join(wdir, "work")},
		{"LOGFILENAME", "org.lsst.ccs.utilites.logging.FileHandlerN.pattern=%W/logs/ccs-logs-%A-testbenchLPC.log"},
	} {
		err := os.Setenv(v[0], v[1])
		if err != nil {
			fatalf("error calling os.Setenv(%q, %q): %v\n", v[0], v[1], err)
		}
	}
}

func dispatch(errc chan error) {
	switch flag.Arg(0) {
	case "lpc":
		runTestbench(errc)
	case "console":
		runConsole(errc)
	case "jas3":
		runJAS3(errc)
	default:
		fatalf("unknown command [%s]\n", flag.Arg(0))
	}
}

func runTestbench(errc chan error) {
	cmd := exec.Command(
		filepath.Join(
			os.Getenv("TEST_ENV_ROOT"),
			"org-lsst-ccs-subsystem-fcs-main-"+fcsVersion,
			"bin",
			"CCSbootstrap.sh",
		),
		"-app", "FcsSubsystem",
		"--description", "/org/lsst/ccs/subsystems/fcs/conf/"+os.Getenv("DESCRIPTION_FILE"),
		"-D", os.Getenv("P0"),
		"-D", os.Getenv("P1"),
		"-D", os.Getenv("P2"),
		"-D", os.Getenv("P3"),
		"-D", os.Getenv("P4"),
		"-D", os.Getenv("P5"),
		"-D", os.Getenv("WORKDIR"),
		"-D", os.Getenv("LOGFILENAME"),
	)
	f, err := os.Create("lpc.log.txt")
	if err != nil {
		log.Printf("could not create logfile: %v\n", err)
		errc <- err
		return
	}
	pr, pw, err := os.Pipe()
	if err != nil {
		log.Printf("could not create pipe: %v\n", err)
		errc <- err
		return
	}

	atexit(func() {
		pr.Sync()
		pw.Sync()
		f.Sync()

		pr.Close()
		pw.Close()
		f.Close()
	})

	stdout := io.MultiWriter(pw, os.Stdout)
	stderr := io.MultiWriter(pw, os.Stderr)

	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	atexit(func() {
		killProc(cmd)
	})

	go func() {
		scan := bufio.NewScanner(pr)
		hdr := []byte("::LPC:: ")
		for scan.Scan() {
			line := scan.Bytes()
			if !bytes.HasPrefix(line, hdr) {
				continue
			}
			f.Write(line[len(hdr):])
			f.Write([]byte("\n"))
		}
	}()
	errc <- cmd.Run()
}

func runConsole(errc chan error) {
	cmd := exec.Command(
		filepath.Join(
			os.Getenv("TEST_ENV_ROOT"),
			"org-lsst-ccs-subsystem-fcs-gui-"+fcsVersion,
			"bin",
			"CCSbootstrap.sh",
		),
		"-app", "ShellCommandConsole",
	)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	atexit(func() {
		killProc(cmd)
	})

	errc <- cmd.Run()
}

func runJAS3(errc chan error) {
	cmd := exec.Command(
		filepath.Join(
			os.Getenv("TEST_ENV_ROOT"),
			"org-lsst-ccs-subsystem-fcs-gui-"+fcsVersion,
			"bin",
			"CCSbootstrap.sh",
		),
		"-app", "CCS-Console",
	)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	atexit(func() {
		killProc(cmd)
	})

	errc <- cmd.Run()
}

func unzip(dest, src string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			fatalf("error closing zip archive: %v\n", err)
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
				fatalf("error closing zip file: %v\n", err)
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
					fatalf("error closing zip output file: %v\n", err)
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
