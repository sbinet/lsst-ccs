package canbus

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
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

func (cmd Command) bytes() []byte {
	o := make([]byte, 0, len(cmd.Name)+1+len(cmd.Data))
	o = append(o, []byte(cmd.Name)...)
	o = append(o, sepComma...)
	o = append(o, cmd.Data...)
	if !bytes.HasSuffix(o, []byte("\n")) {
		o = append(o, []byte("\n")...)
	}
	return o
}

func (cmd Command) String() string {
	return fmt.Sprintf("Command{%s,%s}", cmd.Name, string(cmd.Data))
}

func newCommand(data []byte) Command {
	data = bytes.TrimSpace(data)
	if !bytes.Contains(data, sepComma) {
		return Command{}
	}

	tokens := bytes.SplitN(data, sepComma, 2)
	cmd := Command{
		Name: Cmd(tokens[0]),
		Data: tokens[1],
	}
	return cmd
}

type Bus struct {
	*fwk.Base
	port  int
	l     net.Listener
	conn  net.Conn
	nodes []int

	adc *ADC
	dac *DAC

	devices []fwk.Device
	Send    chan Command
	Recv    chan Command
}

func New(name string, port int, adc *ADC, dac *DAC, devices ...fwk.Device) *Bus {
	devs := append([]fwk.Device{adc, dac}, devices...)
	bus := &Bus{
		Base:    fwk.NewBase(name),
		port:    port,
		nodes:   make([]int, 0, 2),
		adc:     adc,
		dac:     dac,
		Send:    make(chan Command),
		Recv:    make(chan Command),
		devices: devs,
	}
	fwk.System.Register(bus)
	for _, dev := range bus.devices {
		fwk.System.Register(dev)
	}

	return bus
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

	bus.Infof(">>> boot... [done]\n")
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
	if true {
		errc := make(chan error)
		go bus.startCWrapper(errc)
	}

	bus.Infof("... starting tcp server ...\n")
	bus.l, err = net.Listen("tcp", fmt.Sprintf(":%d", bus.port))
	if err != nil {
		bus.Errorf("error starting tcp server: %v\n", err)
		return err
	}

	bus.Infof("... waiting for a connection ...\n")
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
	bus.Infof("handle...\n")
	const bufsz = 1024
	buf := make([]byte, bufsz)

	// consume welcome message
	n, err := bus.conn.Read(buf)
	if err != nil {
		bus.Errorf("error receiving welcome message: %v\n", err)
		return
	}
	if n <= 0 {
		bus.Errorf("empty welcome message!\n")
		return
	}

	if !bytes.HasPrefix(buf[:n], []byte("TestBench ISO-8859-1")) {
		bus.Errorf("unexpected welcome message: %q\n", string(buf[:n]))
		return
	}

	// discover nodes
	for len(bus.nodes) < len(bus.devices) {
		buf = buf[:bufsz]
		n, err := bus.conn.Read(buf)
		if err != nil {
			bus.Errorf("error receiving boot message: %v\n", err)
			return
		}
		if n <= 0 {
			// nothing was read...
			continue
		}
		buf = buf[:n]
		cmd := newCommand(buf)
		switch cmd.Name {
		case Boot:
			id, err := strconv.Atoi(string(cmd.Data))
			if err != nil {
				bus.Errorf("error decoding node id: %v\n", err)
				return
			}
			bus.nodes = append(bus.nodes, id)
		default:
			bus.Errorf("unexpected command name: %q (cmd=%v)\n", cmd.Name, cmd)
		}
	}

	type Node struct {
		id       int
		device   int
		vendor   int
		product  int
		revision int
		serial   string
	}

	nodes := make([]Node, len(bus.nodes))
	// fetch infos about nodes
	for _, id := range bus.nodes {
		buf := []byte(fmt.Sprintf("%s,%d\n", Info, id))
		_, err := bus.conn.Write(buf)
		if err != nil {
			bus.Errorf("error sending info message: %v\n", err)
			return
		}

		buf = make([]byte, bufsz)
		n, err := bus.conn.Read(buf)
		if err != nil {
			bus.Errorf("error receiving info message: %v\n", err)
			return
		}
		if n <= 0 {
			// nothing was read...
			continue
		}
		buf = buf[:n]
		cmd := newCommand(buf)
		switch cmd.Name {
		case Info:
			var node Node
			_, err = fmt.Fscanf(
				bytes.NewReader(cmd.Data),
				"%d,%d,%d,%d,%d,%s",
				&node.id,
				&node.device,
				&node.vendor,
				&node.product,
				&node.revision,
				&node.serial,
			)
			if err != nil {
				bus.Errorf("error decoding %v: %v\n", cmd, err)
				return
			}
			bus.Infof("node=%v\n", node)
			nodes = append(nodes, node)
			//TODO(sbinet): better/more-general handling
			switch node.serial {
			case bus.adc.serial:
				bus.adc.node = node.id
				bus.adc.bus = bus
			case bus.dac.serial:
				bus.dac.node = node.id
				bus.dac.bus = bus
			}

		default:
			bus.Errorf("unexpected command name: %q (cmd: %v)\n", cmd.Name, cmd)
			return
		}
	}

	bus.Infof("adc=%#v\n", bus.adc)
	bus.Infof("dac=%#v\n", bus.dac)

loop:
	for {
		select {
		case cmd := <-bus.Send:
			n, err := bus.conn.Write(cmd.bytes())
			if err != nil {
				bus.Errorf("error sending command %v: %v\n", cmd, err)
				return
			}

			// TODO(sbinet) only read back when needed?
			buf = buf[:bufsz]
			n, err = bus.conn.Read(buf)
			if err != nil {
				bus.Errorf("error receiving message: %v\n", err)
				return
			}
			buf = buf[:n]
			cmd = newCommand(buf)
			bus.Recv <- cmd

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
