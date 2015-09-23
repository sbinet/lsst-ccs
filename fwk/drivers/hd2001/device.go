package hd2001

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/sbinet/lsst-ccs/fwk"
	"github.com/sbinet/lsst-ccs/fwk/drivers/canbus"
	"golang.org/x/net/context"
)

type Device struct {
	*fwk.Base
	bus *canbus.Bus

	booted bool
	node   int32
	major  int32
	minor  int32
}

func (dev *Device) Run(ctx context.Context) error {
	var err error
	return err
}

func (dev *Device) Boot(ctx context.Context) error {
	dev.Infof(">>> boot...\n")
	//var err error
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	dev.Infof("--- booting bus...\n")
	go dev.bus.Boot(ctx)

	dev.Infof("for-select...\n")

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case cmd := <-dev.bus.Recv:
			dev.Infof("received: %v\n", cmd)
			switch cmd.Name {
			case canbus.Boot:
				id, err := strconv.Atoi(string(cmd.Data))
				if err != nil {
					dev.Errorf("error decoding node id: %v\n", err)
					return err
				}
				dev.Infof("node=%d\n", id)
				dev.node = int32(id)
				dev.booted = true

				reply := canbus.Command{
					Name: canbus.Info,
					Data: []byte(fmt.Sprintf("%d", id)),
				}

				dev.bus.Send <- reply
				dev.Infof("sent reply: %v\n", reply)

			case canbus.Info:
				dev.Infof("info: %v\n", string(cmd.Data))
			}
		}
	}
	return ctx.Err()
}

func (dev *Device) Start(ctx context.Context) error {
	var err error
	return err
}

func (dev *Device) Stop(ctx context.Context) error {
	var err error
	return err
}

func (dev *Device) Shutdown(ctx context.Context) error {
	var err error
	return err
}

func New(name string, port int) *Device {
	return &Device{
		Base: fwk.NewBase(name),
		bus:  canbus.New("canbus-"+name, port),
	}
}

func (dev *Device) Temperature() float64 {
	dev.bus.Send <- canbus.Command{
		Name: canbus.Rsdo,
		Data: []byte(fmt.Sprintf("%d,%d,%d", dev.node, dev.major, dev.minor)),
	}
	cmd := <-dev.bus.Recv
	temp := 0.0
	fmt.Fscanf(bytes.NewReader(cmd.Data), "%f", &temp)
	return temp
}
