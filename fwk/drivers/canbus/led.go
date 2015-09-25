package canbus

import (
	"fmt"
	"time"

	"github.com/sbinet/lsst-ccs/fwk"
	"golang.org/x/net/context"
)

type LED struct {
	*fwk.Base
	bus *Bus
	dac *DAC

	cid uint8 // channel index on DAC
}

func (led *LED) Start(ctx context.Context) error {
	led.Infof(">>> boot...\n")
	led.dac = led.bus.DAC()
	return nil
}

func (led *LED) Stop(ctx context.Context) error {
	var err error
	err = led.TurnOff()
	if err != nil {
		return err
	}

	return err
}

func (led *LED) Tick(ctx context.Context) error {
	led.Debugf("tick...\n")
	var err error

	err = led.TurnOn()
	if err != nil {
		return err
	}

	time.Sleep(500 * time.Millisecond)

	err = led.TurnOff()
	if err != nil {
		return err
	}
	return err
}

func (led *LED) TurnOn() error {
	return led.write(0x14000)
}

func (led *LED) TurnOff() error {
	return led.write(0x0)
}

func (led *LED) write(value uint32) error {
	var err error
	const subchannel = 0x2
	cmd, err := led.bus.Send(Command{
		Name: Wsdo,
		Data: []byte(fmt.Sprintf("%x,%x,%x,%x,%x",
			led.dac.Node(),
			0x6411,
			led.cid,
			subchannel,
			value,
		)),
	})
	if err != nil {
		return err
	}
	return cmd.Err()
	/*
		cmd = <-led.bus.recv
		switch cmd.Name {
		case Wsdo:
			node := 0
			ecode := 0
			n, err := fmt.Fscanf(
				bytes.NewReader(cmd.Data),
				"%x,%x",
				&ecode,
				&node,
			)
			if err != nil {
				return err
			}
			if n <= 0 {
				return io.ErrShortBuffer
			}
			if node != led.dac.Node() {
				return fmt.Errorf("unexpected node. got=0x%x want=0x%x", node,
					led.dac.Node())
			}
		default:
			return fmt.Errorf("unexpected command %v", cmd)
		}

		//led.bus.Send <- Command{Rsdo, []byte(fmt.Sprintf("%x", led.dac.Node()))}
		//cmd = <-led.bus.Recv
		//led.Infof("cmd-sync: %v\n", cmd)
		return err
	*/
}

func NewLED(name string, bus string) *LED {
	busdev := fwk.System.Device(bus)
	led := &LED{
		Base: fwk.NewBase(name),
		bus:  busdev.(*Bus),
		cid:  0x1,
	}
	return led
}
