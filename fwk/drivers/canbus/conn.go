package canbus

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"

	"github.com/gonuts/logger"
)

// cwrapper manages the connection to the C-Wrapper program
type cwrapper struct {
	port int
	lst  net.Listener
	conn net.Conn
}

func (c *cwrapper) init(msg *logger.Logger) error {
	var err error

	if true {
		errc := make(chan error)
		go c.startCWrapper(msg, errc)
	}

	msg.Infof("... starting tcp server ...\n")
	c.lst, err = net.Listen("tcp", fmt.Sprintf(":%d", c.port))
	if err != nil {
		msg.Errorf("error starting tcp server: %v\n", err)
		return err
	}

	msg.Infof("... waiting for a connection ...\n")
	c.conn, err = c.lst.Accept()
	if err != nil {
		msg.Errorf("error accepting connection: %v\n", err)
		return err
	}

	return err
}

func (c *cwrapper) Read(data []byte) (int, error) {
	return c.conn.Read(data)
}

func (c *cwrapper) Write(data []byte) (int, error) {
	return c.conn.Write(data)
}

func (c *cwrapper) Close() error {
	var err error
	if c.conn != nil {
		errConn := c.conn.Close()
		if errConn != nil {
			err = errConn
		}

	}

	if c.lst != nil {
		errLst := c.lst.Close()
		if errLst != nil {
			err = errLst
		}
	}

	c.lst = nil
	c.conn = nil
	return err
}

func (c *cwrapper) startCWrapper(msg *logger.Logger, errc chan error) {
	host, err := c.host(msg)
	if err != nil {
		errc <- err
		return
	}

	msg.Infof("Starting c-wrapper on PC-104... (listen for %s:%d)\n", host, c.port)

	cmd := exec.Command(
		"ssh",
		"-X",
		"root@clrlsstemb01.in2p3.fr",
		"startCWrapper --host="+host, fmt.Sprintf("--port=%d", c.port),
	)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "TERM=vt100")
	//cmd.Stdin = os.Stdin
	//cmd.Stdout = os.Stderr
	//cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	/*
		atexit(func() {
			killProc(cmd)
		})
	*/

	msg.Infof("c-wrapper command: %v\n", cmd.Args)

	err = cmd.Run()
	if err != nil {
		msg.Errorf("c-wrapper= %v\n", err)
		errc <- err
		return
	}
}

func (c *cwrapper) host(msg *logger.Logger) (string, error) {
	host, err := os.Hostname()
	if err != nil {
		msg.Errorf("could not retrieve hostname: %v\n", err)
		return "", err
	}

	addrs, err := net.LookupIP(host)
	if err != nil {
		msg.Errorf("could not lookup hostname IP: %v\n", err)
		return "", err
	}

	for _, addr := range addrs {
		ipv4 := addr.To4()
		if ipv4 == nil {
			continue
		}
		return ipv4.String(), nil
	}

	msg.Errorf("could not infer host IP")
	return "", fmt.Errorf("could not infer host IP")
}
