// fcs-run runs a command within the CCS environment.
package main

import (
	"archive/zip"
	"encoding/xml"
	"flag"
	"fmt"
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
	"time"
)

var (
	haddr    = flag.String("addr", "", "<ip>[:<port>] PC-104 will listen to")
	cwrapper = flag.Bool("cwrapper", false, "start c-wrapper")

	host = ""
	port = 50000

	exitFuncs  = []func(){}
	wdir       = "."
	fcsVersion = "dev-SNAPSHOT"
	dbVersion  = "dev-SNAPSHOT"
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

	//go testNet()
	run()
}

func run() {
	var err error
	wdir, err = os.Getwd()
	if err != nil {
		fatalf("could not retrieve current work directory: %v\n", wdir)
	}

	errc := make(chan error)

	initProject()
	//makeDistrib()
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
		"startCWrapper --host="+host, fmt.Sprintf("--port=%d", port),
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
	for _, proj := range []struct {
		Dir     string
		Version *string
	}{
		{
			Dir:     "org-lsst-ccs-subsystem-fcs",
			Version: &fcsVersion,
		},
		{
			Dir:     "org-lsst-ccs-localdb",
			Version: &dbVersion,
		},
	} {

		fname := filepath.Join(wdir, proj.Dir, "pom.xml")
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
		*proj.Version = data.Version
	}
}

func setupEnv() {
	for _, v := range [][2]string{
		{"FCS_VERSION", fcsVersion},
		{"LOCALDB_VERSION", dbVersion},
		// logging
		{"P0", "java.util.logging.ConsoleHandler.level=ALL"},
		{"P1", "java.util.logging.FileHandler.formatter=org.lsst.ccs.utilities.logging.TextFormatter"},
		{"P2", ".level=WARNING"},
		{"P3", "org.lsst.ccs.level=ALL"},
		{"P4", "org.lsst.ccs.bus.level=INFO"},
		{"P5", "org.lsst.ccs.subsystems.fcs.level=ALL"},

		{"TEST_ENV_ROOT", filepath.Join(wdir, "DISTRIB")},
		{"TEST_ENV_BIN", filepath.Join("${TEST_ENV_ROOT}", "bin")},
		{"TEST_ENV_RESOURCES", filepath.Join("${TEST_ENV_ROOT}", "externalResources")},
		{"TEST_ENV_DRIVERS", filepath.Join("${TEST_ENV_ROOT}", "drivers")},

		// needed by CCSbootstrap.sh
		{"CCS_RESOURCE_PATH", "${TEST_ENV_RESOURCES}"},

		// make sure we use IPv4 in JAS
		{"JASJVM_OPTS", "-Djava.net.preferIPV4Stack=true"},

		{"DESCRIPTION_FILE", "testbenchLPC.groo"},
		{"CONFIGURATION_FILE", "testbenchLPC_XXXXX_.properties"},
		{"WORKDIR", filepath.Join(wdir, "work")},
		{"LOGFILENAME", "java.util.logging.FileHandler.pattern=%W/logs/ccs-logs-%A-testbenchLPC.log"},
	} {
		err := os.Setenv(v[0], os.ExpandEnv(v[1]))
		if err != nil {
			fatalf("error calling os.Setenv(%q, %q): %v\n", v[0], v[1], err)
		}
	}
}

func dispatch(errc chan error) {
	switch flag.Arg(0) {
	case "start-localdb", "jas3", "list", "infos", "shell":
		// ok
	default:
		*cwrapper = true
	}

	if *cwrapper {
		go startCWrapper(errc)
	}

	switch flag.Arg(0) {
	case "lpc":
		runTestbench(errc)
	case "sim-autochanger":
		runSimAutochanger(errc)
	case "console":
		runConsole(errc)
	case "jas3":
		runJAS3(errc)
	case "list":
		runListApps(errc)
	case "infos":
		runInfos(errc)
	case "start-localdb":
		startLocalDB(errc)
	case "shell":
		runShell(errc)
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
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("running: %v...\n", cmd.Args)

	atexit(func() {
		killProc(cmd)
	})

	errc <- cmd.Run()
}

func runSimAutochanger(errc chan error) {
	os.Setenv("DESCRIPTION_FILE", "autochanger__simulation.groovy")
	os.Setenv("CONFIGURATION_FILE", "")

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
		"-D", "org.lsst.ccs.startInEngineeringMode=true",
		"-D", "org.lsst.ccs.logging.StackTraceFormats.depth=-1",
		"-D", os.Getenv("WORKDIR"),
		"-D", os.Getenv("LOGFILENAME"),
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("running: %v...\n", cmd.Args)

	atexit(func() {
		killProc(cmd)
	})

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

func runListApps(errc chan error) {
	cmd := exec.Command(
		filepath.Join(
			os.Getenv("TEST_ENV_ROOT"),
			"org-lsst-ccs-subsystem-fcs-main-"+fcsVersion,
			"bin",
			"CCSbootstrap.sh",
		),
		"-la",
	)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	atexit(func() {
		killProc(cmd)
	})

	errc <- cmd.Run()
}

func startLocalDB(errc chan error) {
	cmd := exec.Command(
		filepath.Join(
			os.Getenv("TEST_ENV_ROOT"),
			"org-lsst-ccs-localdb-main-"+dbVersion,
			"bin",
			"CCSbootstrap.sh",
		),
		"-app", "TrendingIngestModule",
		"-D",
		"org.lsst.ccs.localdb.hibernate.properties.file="+filepath.Join(
			os.Getenv("TEST_ENV_ROOT"),
			"org-lsst-ccs-subsystem-fcs-main-"+fcsVersion,
			"etc",
			"statusPersister.properties",
		),
	)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	atexit(func() {
		killProc(cmd)
	})

	errc <- cmd.Run()
}

func runInfos(errc chan error) {
	cmd := exec.Command(
		filepath.Join(
			os.Getenv("TEST_ENV_ROOT"),
			"org-lsst-ccs-subsystem-fcs-main-"+fcsVersion,
			"bin",
			"CCSbootstrap.sh",
		),
		"-distInfo",
	)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	atexit(func() {
		killProc(cmd)
	})

	errc <- cmd.Run()
}

func runShell(errc chan error) {
	cmd := exec.Command("/bin/bash")
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

func testNet() {
	srv, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatalf("error listening: %v\n", err)
	}

	for {
		conn, err := srv.Accept()
		if err != nil {
			log.Fatalf("error accept: %v\n", err)
		}
		go testForward(conn)
	}
}

func testForward(conn net.Conn) {
	time.Sleep(20 * time.Second)

	cli, err := net.DialTimeout(
		"tcp", "134.158.120.94:50000",
		1000*time.Second,
	)
	if err != nil {
		log.Fatalf("error dialing: %v\n", err)
	}

	fw, err := os.Create("sock-w.txt")
	if err != nil {
		log.Fatalf("error fw: %v\n", err)
	}

	w := io.MultiWriter(fw, cli)
	go func() {
		defer fw.Close()
		io.Copy(w, conn)
	}()

	r := io.TeeReader(cli, fw)
	go io.Copy(conn, r)
}
