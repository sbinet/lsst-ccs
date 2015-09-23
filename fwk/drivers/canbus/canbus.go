package canbus

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/net/context"

	"github.com/sbinet/lsst-ccs/fwk"
)

// Cmd is a type of command to send/receive on/from the CAN bus.
type Cmd string

// The command types known to the CAN bus.
const (
	Boot Cmd = "boot"
	Info     = "info"
	Rsdo     = "rsdo"
	Wsdo     = "wsdo"
	Sync     = "sync"
)

// Command is a command sent/received on the CAN bus
type Command struct {
	Name Cmd
	Data []byte
}

func (cmd *Command) bytes() []byte {
	o := make([]byte, 0, len(cmd.Name)+1+len(cmd.Data))
	o = append(o, []byte(cmd.Name)...)
	o = append(o, sepComma...)
	o = append(o, cmd.Data...)
	if !bytes.HasSuffix(o, []byte("\n")) {
		o = append(o, []byte("\n")...)
	}
	return o
}

type Bus struct {
	*fwk.Base
	port  int
	l     net.Listener
	conn  net.Conn
	nodes []int

	Send chan Command
	Recv chan Command
}

func New(name string, port int) *Bus {
	return &Bus{
		Base:  fwk.NewBase(name),
		port:  port,
		nodes: make([]int, 0, 2),
		Send:  make(chan Command),
		Recv:  make(chan Command),
	}
}

func (bus *Bus) Run(ctx context.Context) error {
	var err error
	return err
}

func (bus *Bus) Boot(ctx context.Context) error {
	bus.Infof(">>> boot...\n")
	var err error
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err = bus.init()
	if err != nil {
		bus.Errorf("error: %v\n", err)
		return err
	}

	return err
}

func (bus *Bus) Start(ctx context.Context) error {
	var err error
	return err
}

func (bus *Bus) Stop(ctx context.Context) error {
	var err error
	return err
}

func (bus *Bus) Shutdown(ctx context.Context) error {
	var err error
	return err
}

func (bus *Bus) init() error {
	var err error
	errc := make(chan error)
	go bus.startCWrapper(errc)

	bus.l, err = net.Listen("tcp", fmt.Sprintf(":%d", bus.port))
	if err != nil {
		bus.Errorf("error starting tcp server: %v\n", err)
		return err
	}

	bus.conn, err = bus.l.Accept()
	if err != nil {
		bus.Errorf("error accepting connection: %v\n", err)
		return err
	}

	go bus.handle(nil)

	return err
}

func (bus *Bus) Close() error {
	if bus.l == nil {
		return nil
	}
	return bus.l.Close()
}

func (bus *Bus) handle(quit chan struct{}) {

	reader := chanFromConn(bus.conn)
	writer := chanToConn(bus.conn)

loop:
	for {
		select {
		case msg := <-reader:
			if msg.err != nil {
				bus.Errorf("error receiving message: %v\n", msg.err)
				return
			}

			msg.data = bytes.TrimSpace(msg.data)
			if !bytes.Contains(msg.data, sepComma) {
				bus.Debugf("received: %q\n", string(msg.data))
				continue
			}

			tokens := bytes.SplitN(msg.data, sepComma, 2)
			cmd := Command{
				Name: Cmd(tokens[0]),
				Data: tokens[1],
			}
			bus.Infof("tokens: %v\n", cmd)
			/*
				switch cmd.Name {
				case Boot:
					id, err := strconv.Atoi(string(cmd.Data))
					if err != nil {
						bus.Errorf("error decoding node id: %v\n", err)
						continue
					}
					bus.nodes = append(bus.nodes, id)
					go func() {
						writer <- message{
							data: []byte(fmt.Sprintf("%s,%d", Info, id)),
						}
					}()
					continue
				}
			*/
			bus.Recv <- cmd

		case cmd := <-bus.Send:
			go func() {
				writer <- message{
					data: cmd.bytes(),
				}
			}()

		case <-quit:
			bus.Debugf("quit...\n")
			break loop
		}
	}

}

var (
	sepComma = []byte(",")
)

type message struct {
	data []byte
	err  error
}

// chanFromConn creates a channel from a Conn object, and sends everything it
//  Read()s from the socket to the channel.
func chanFromConn(conn net.Conn) chan message {
	c := make(chan message)

	go func() {
		b := make([]byte, 1024)

		for {
			n, err := conn.Read(b)
			if n <= 0 {
				c <- message{data: nil, err: err}
				break
			}
			res := make([]byte, n)
			// Copy the buffer so it doesn't get changed while read by the recipient.
			copy(res, b[:n])
			c <- message{data: res, err: err}
		}
	}()

	return c
}

func chanToConn(conn net.Conn) chan message {
	c := make(chan message)
	go func() {
		for {
			msg := <-c
			_, err := io.Copy(conn, bytes.NewReader(msg.data))
			//_, err := conn.Write(msg.data)
			if err != nil {
				break
			}
		}
	}()
	return c
}

func (bus *Bus) startCWrapper(errc chan error) {
	host, err := bus.host()
	if err != nil {
		errc <- err
		return
	}

	bus.Infof("Starting c-wrapper on PC-104... (listen for %s:%d)\n", host, bus.port)

	cmd := exec.Command(
		"ssh",
		"-X",
		"root@clrlsstemb01.in2p3.fr",
		"startCWrapper --host="+host, fmt.Sprintf("--port=%d", bus.port),
	)
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

	bus.Infof("c-wrapper command: %v\n", cmd.Args)

	err = cmd.Run()
	if err != nil {
		bus.Errorf("c-wrapper= %v\n", err)
		errc <- err
		return
	}
}

func (bus *Bus) host() (string, error) {
	host, err := os.Hostname()
	if err != nil {
		bus.Errorf("could not retrieve hostname: %v\n", err)
		return "", err
	}

	addrs, err := net.LookupIP(host)
	if err != nil {
		bus.Errorf("could not lookup hostname IP: %v\n", err)
		return "", err
	}

	for _, addr := range addrs {
		ipv4 := addr.To4()
		if ipv4 == nil {
			continue
		}
		return ipv4.String(), nil
	}

	bus.Errorf("could not infer host IP")
	return "", fmt.Errorf("could not infer host IP")
}
