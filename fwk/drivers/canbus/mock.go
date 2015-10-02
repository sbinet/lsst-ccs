package canbus

import (
	"bytes"
	"fmt"

	"github.com/sbinet/lsst-ccs/fwk"
	"golang.org/x/net/context"
)

// mockBus is canbus mock
type mockBus struct {
	*fwk.Base
	port  int
	nodes []int

	adc *ADC
	dac *DAC

	devices []fwk.Device
}

func NewMock(name string, port int, adc *ADC, dac *DAC, devices ...fwk.Device) fwk.Module {
	devs := append([]fwk.Device{adc, dac}, devices...)
	bus := &mockBus{
		Base:    fwk.NewBase(name),
		port:    port,
		adc:     adc,
		dac:     dac,
		devices: devs,
	}
	fwk.System.Register(bus)
	for _, dev := range bus.devices {
		fwk.System.Register(dev)
	}

	return bus
}

func (bus *mockBus) Boot(ctx context.Context) error {
	var err error
	// adc is 0x41
	bus.adc.node = 0x41
	bus.adc.bus = bus

	// dac is 0x??
	bus.dac.node = 0x42
	bus.dac.bus = bus

	bus.Infof("adc=%#v\n", bus.adc)
	bus.Infof("dac=%#v\n", bus.dac)

	err = bus.adc.init()
	if err != nil {
		bus.Errorf("error initializing ADC: %v\n", err)
		return err
	}

	err = bus.dac.init()
	if err != nil {
		bus.Errorf("error initializing DAC: %v\n", err)
		return err
	}

	//go bus.run()

	return err
}

func (bus *mockBus) Start(ctx context.Context) error {
	var err error
	return err
}

func (bus *mockBus) Stop(ctx context.Context) error {
	var err error
	return err
}

func (bus *mockBus) Shutdown(ctx context.Context) error {
	var err error
	bus.Infof("shutdown...\n")

	_, err = bus.Send(Command{Quit, nil})
	if err != nil {
		bus.Errorf("error closing canbus: %v\n", err)
		return err
	}

	return err
}

func (bus *mockBus) ADC() *ADC {
	return bus.adc
}

func (bus *mockBus) DAC() *DAC {
	return bus.dac
}

func (bus *mockBus) Send(icmd Command) (Command, error) {
	bus.Debugf("request: %v\n", icmd)
	var ocmd Command
	var err error

	switch icmd.Name {
	case Quit:
		return icmd, err

	case Rsdo:
		var node int
		var idx int
		var sub int
		_, err = fmt.Fscanf(bytes.NewReader(icmd.Data),
			"%x,%x,%x",
			&node,
			&idx,
			&sub,
		)
		if err != nil {
			return ocmd, err
		}
		return Command{
			Name: Rsdo,
			Data: []byte(fmt.Sprintf(
				"%x,%x,%x",
				node,
				0,
				(2<<14)/2,
			)),
		}, err

	case Wsdo:
		var node int
		var idx int
		_, err = fmt.Fscanf(bytes.NewReader(icmd.Data),
			"%x,%x",
			&node,
			&idx,
		)
		if err != nil {
			return ocmd, err
		}

		return Command{
			Name: Wsdo,
			Data: []byte(fmt.Sprintf(
				"%x,%x",
				node,
				0,
			)),
		}, err
	}

	return ocmd, err
}
