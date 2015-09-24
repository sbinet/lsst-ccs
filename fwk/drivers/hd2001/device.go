package hd2001

import (
	"bytes"
	"fmt"

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

	dev.Infof("for-select...\n")

	if true {
		go func() {
			dev.bus.Send <- canbus.Command{
				Name: canbus.Rsdo,
				Data: []byte("41,6401,3"),
			}
			dev.Infof("sent 3\n")
		}()
		go func() {
			dev.bus.Send <- canbus.Command{
				Name: canbus.Rsdo,
				Data: []byte("41,6401,2"),
			}
			dev.Infof("sent 2\n")
		}()
		go func() {
			dev.bus.Send <- canbus.Command{
				Name: canbus.Rsdo,
				Data: []byte("41,6401,4"),
			}
			dev.Infof("sent 4\n")
		}()

	}
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case cmd := <-dev.bus.Recv:
			dev.Infof("received: %v\n", cmd)
			switch cmd.Name {
			case canbus.Rsdo:
				dev.Infof("rsdo: %v\n", cmd)

			default:
				dev.Errorf("unknown canbus.Cmd: %v\n", cmd.Name)
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

func New(name string, bus string) *Device {
	dev := fwk.System.Device(bus)
	return &Device{
		Base: fwk.NewBase(name),
		bus:  dev.(*canbus.Bus),
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
