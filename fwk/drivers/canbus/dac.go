package canbus

import (
	"github.com/sbinet/lsst-ccs/fwk"
)

type DAC struct {
	*fwk.Base
	node   int
	serial string
	bus    *Bus
}

func (dac *DAC) Node() int {
	return dac.node
}

func (dac *DAC) Serial() string {
	return dac.serial
}

func (dac *DAC) init() error {
	var err error
	return err
}

func NewDAC(name, serial string) *DAC {
	return &DAC{
		Base:   fwk.NewBase(name),
		serial: serial,
	}
}
