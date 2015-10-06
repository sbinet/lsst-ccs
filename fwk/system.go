package fwk

import (
	"fmt"

	"github.com/gdamore/mangos"
	"github.com/gdamore/mangos/protocol/bus"
	"github.com/gdamore/mangos/transport/ipc"
	"github.com/gdamore/mangos/transport/tcp"
)

const (
	// BusAddr is the default rendez-vous point for the system bus
	BusAddr = "tcp://127.0.0.1:40000"
)

var System = systemType{
	name:    "root",
	devices: make([]Device, 0, 2),
	devmap:  make(map[string]Device),
}

type systemType struct {
	name    string
	devices []Device
	devmap  map[string]Device

	sock mangos.Socket
}

func (sys *systemType) Devices() []Device {
	return sys.devices
}

func (sys *systemType) Name() string {
	return sys.name
}

func (sys *systemType) init() error {
	sock, err := bus.NewSocket()
	if err != nil {
		return err
	}
	sys.sock = sock
	sys.sock.AddTransport(ipc.NewTransport())
	sys.sock.AddTransport(tcp.NewTransport())

	err = sys.sock.Listen(BusAddr)
	if err != nil {
		return err
	}

	return err
}

func (sys *systemType) Register(dev Device) {
	sys.devices = append(sys.devices, dev)
	d, dup := sys.devmap[dev.Name()]
	if dup {
		panic(fmt.Errorf(
			"fwk: duplicate device %q\nold=%#v\nnew=%#v",
			dev.Name(),
			d, dev,
		))
	}
	sys.devmap[dev.Name()] = dev
}

func (sys *systemType) Device(name string) Device {
	return sys.devmap[name]
}
