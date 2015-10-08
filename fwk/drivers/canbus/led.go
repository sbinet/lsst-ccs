package canbus

import (
	"fmt"
	"time"

	"github.com/sbinet/lsst-ccs/fwk"
	"golang.org/x/net/context"
)

type LED struct {
	*fwk.Base
	bus Bus
	dac *DAC

	cid uint8 // channel index on DAC
}

func (led *LED) Boot(ctx context.Context) error {
	var err error
	led.dac = led.bus.DAC()
	return err
}

func (led *LED) Start(ctx context.Context) error {
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

func (led *LED) Shutdown(ctx context.Context) error {
	return nil
}

func (led *LED) Tick(ctx context.Context) error {
	led.Debugf("tick...\n")
	var err error

	err = led.TurnOn()
	if err != nil {
		led.Errorf("error turning LED ON: %v\n", err)
		return err
	}

	time.Sleep(500 * time.Millisecond)

	err = led.TurnOff()
	if err != nil {
		led.Errorf("error turning LED OFF: %v\n", err)
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
	const subchannel = 0x2
	msg := Msg(Wsdo, []byte(fmt.Sprintf(
		"%x,%x,%x,%x,%x",
		led.dac.Node(),
		0x6411,
		led.cid,
		subchannel,
		value,
	)))
	led.Infof("--> %q...\n", string(msg.Req.Data))
	led.bus.Queue() <- msg
	cmd := <-msg.Reply
	led.Infof("<-- %q... | %q\n",
		string(msg.Req.Data),
		string(cmd.Data),
	)
	return cmd.Err()
}

func NewLED(name string, bus string) *LED {
	busdev := fwk.System.Device(bus)
	led := &LED{
		Base: fwk.NewBase(name),
		bus:  busdev.(Bus),
		cid:  0x1,
	}
	return led
}
