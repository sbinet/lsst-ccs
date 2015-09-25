package canbus

import (
	"bytes"
	"fmt"
	"io"

	"github.com/sbinet/lsst-ccs/fwk"
)

const (
	adcVoltsPerBit  = 0.3125 * 1e-3 // in Volts
	waterFreezeTemp = 273.15
)

type ADC struct {
	*fwk.Base
	node   int
	serial string
	tx     int
	bus    *Bus
}

func (adc *ADC) Node() int {
	return adc.node
}

func (adc *ADC) Serial() string {
	return adc.serial
}

func (adc *ADC) Tx() int {
	return adc.tx
}

func (*ADC) Volts(adc int) float64 {
	return float64(adc) * adcVoltsPerBit
}

func (adc *ADC) init() error {
	if true {
		return nil
	}

	const sz = 1 // len of adc.tx
	const subchannel = 2
	for _, channel := range []int{0x1801, 0x1802} {
		n, err := adc.bus.conn.Write(
			Command{
				Name: Wsdo,
				Data: []byte(fmt.Sprintf(
					"%x,%x,%x,%x,%x",
					adc.node, channel, subchannel, sz, adc.tx,
				)),
			}.bytes(),
		)
		if err != nil {
			return err
		}
		if n <= 0 {
			return io.ErrShortWrite
		}

		buf := make([]byte, 1024)
		n, err = adc.bus.conn.Read(buf)
		if err != nil {
			return err
		}
		if n <= 0 {
			return io.ErrShortBuffer
		}
		buf = buf[:n]
		cmd := newCommand(buf)
		adc.Infof("channel: 0x%x => %v\n", channel, cmd)
		switch cmd.Name {
		case Wsdo:
			node := 0
			ecode := 0
			n, err = fmt.Fscanf(
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
			if node != adc.node {
				return fmt.Errorf("unexpected node. got=0x%x want=0x%x", node, adc.node)
			}
			if ecode != 0 {
				return fmt.Errorf("canbus error (ecode=%d)", ecode)
			}
		default:
			return fmt.Errorf("unexpected command %v", cmd)
		}
	}

	return nil
}

func NewADC(name, serial string, tx int) *ADC {
	return &ADC{
		Base:   fwk.NewBase(name),
		serial: serial,
		tx:     tx,
	}
}

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
