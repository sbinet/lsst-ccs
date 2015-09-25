package main

import (
	"github.com/sbinet/lsst-ccs/fwk"
	"github.com/sbinet/lsst-ccs/fwk/drivers/canbus"
	"github.com/sbinet/lsst-ccs/fwk/drivers/hd2001"
)

const (
	port = 50000
)

func main() {

	app, err := fwk.New(
		"lpc",
		canbus.New(
			"canbus", port,
			canbus.NewADC("ai814", "c7c80499", 0x1),
			canbus.NewDAC("ao412", "c7c60327"),
		),
		canbus.NewLED("led", "canbus"),
		hd2001.New("hpt", "canbus"),
	)
	if err != nil {
		panic(err)
	}

	err = app.Run()
	if err != nil {
		panic(err)
	}
}
