package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/sbinet/lsst-ccs/fwk"
	"github.com/sbinet/lsst-ccs/fwk/drivers/hd2001"
)

const (
	port = "50000"
)

var (
	sepComma = []byte(",")
)

func main() {

	app, err := fwk.New(
		"lpc",
		hd2001.New("hpt", 50000),
	)
	if err != nil {
		panic(err)
	}

	err = app.Run()
	if err != nil {
		panic(err)
	}

	// Listen for incoming connections.
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	// Close the listener when the application closes.
	defer listen.Close()
	fmt.Println("Listening on localhost:" + port)
	for {
		// Listen for an incoming connection.
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	hw := newHwMgr(conn)
	hw.run()
	hw.Close()
}

var (
	cmdBoot = "boot"
	cmdInfo = "info"
)

type message struct {
	data []byte
	err  error
}

type command struct {
	name string
	data []byte
}

type hwMgr struct {
	nodes []int
	rcmd  chan command
	wcmd  chan command
	quit  chan struct{}
	conn  net.Conn
}

func (hw *hwMgr) Close() error {
	hw.quit <- struct{}{}
	return hw.conn.Close()
}

func newHwMgr(conn net.Conn) *hwMgr {
	return &hwMgr{
		nodes: make([]int, 0, 2),
		rcmd:  make(chan command),
		wcmd:  make(chan command),
		quit:  make(chan struct{}),
		conn:  conn,
	}
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

func (hw *hwMgr) run() {
	fmt.Printf("hw-run-loop...\n")

	rch := chanFromConn(hw.conn)
	wch := chanToConn(hw.conn)

loop:
	for {
		fmt.Printf("hw-run-select...\n")
		select {
		case msg := <-rch:
			if msg.err != nil {
				fmt.Printf("error: %v\n", msg.err)
				hw.quit <- struct{}{}
				continue
			}

			msg.data = bytes.TrimSpace(msg.data)
			if !bytes.Contains(msg.data, sepComma) {
				fmt.Printf("received: %q\n", string(msg.data))
				continue
			}
			fmt.Printf("received-msg: %q\n", string(msg.data))

			tokens := bytes.SplitN(msg.data, sepComma, 2)
			fmt.Printf("handle-command: %s...\n", string(msg.data))
			go func() {
				hw.rcmd <- command{
					name: string(tokens[0]),
					data: tokens[1],
				}
			}()

		case cmd := <-hw.wcmd:
			fmt.Printf("send-command: %s,%s...\n", cmd.name, string(cmd.data))
			go func() {
				wch <- message{
					data: []byte(fmt.Sprintf("%s,%s\n", cmd.name, string(cmd.data))),
				}
			}()

		case cmd := <-hw.rcmd:
			fmt.Printf(
				"received command: %s - %s\n",
				cmd.name,
				string(cmd.data),
			)
			switch cmd.name {
			case cmdBoot:
				id, err := strconv.Atoi(string(cmd.data))
				if err != nil {
					fmt.Printf("error: %v\n", err)
					continue
				}
				hw.nodes = append(hw.nodes, id)
				go func() {
					//time.Sleep(10 * time.Second)
					hw.wcmd <- command{
						name: "info",
						data: []byte(fmt.Sprintf("%d", id)),
					}
				}()
			case cmdInfo:
			default:
				fmt.Printf("** unknown command: %q\n", cmd.name)
			}
		case <-hw.quit:
			break loop
		}
	}
	fmt.Printf("hw-run-loop... [done]\n")
}
