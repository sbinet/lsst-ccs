// fcs-ana analyzes a fcs-mgr <some-command> output file
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
)

func main() {
	fname := os.Args[1]
	f, err := os.Open(fname)
	if err != nil {
		log.Fatalf("could not open [%s]: %v\n", fname, err)
	}
	defer f.Close()

	raws := make(map[string]plotter.XYs, 3)
	vals := make(map[string]plotter.XYs, 3)

	keys := []string{"temp", "press", "hygro"}
	for _, k := range keys {
		raws[k] = make(plotter.XYs, 0, 1024)
		vals[k] = make(plotter.XYs, 0, 1024)
	}

	i := 0
	scan := bufio.NewScanner(f)
	for scan.Scan() {
		line := scan.Bytes()
		if bytes.HasPrefix(line, []byte("#")) {
			continue
		}

		r := bytes.NewReader(line)
		var evt Event
		err = json.NewDecoder(r).Decode(&evt)
		if err != nil {
			log.Fatalf("error decoding line %q: %v\n", string(line), err)
		}
		raws["temp"] = append(raws["temp"],
			XY{
				X: float64(i * 3), // data is snapshot every 3s
				Y: adcToTemperature(evt.Temp.Raw),
			},
		)
		vals["temp"] = append(vals["temp"],
			XY{
				X: float64(i * 3), // data is snapshot every 3s
				Y: adcToTemperature(evt.Temp.Value),
			},
		)
		raws["press"] = append(raws["press"],
			XY{
				X: float64(i * 3), // data is snapshot every 3s
				Y: adcToPressure(evt.Pressure.Raw),
			},
		)
		vals["press"] = append(vals["press"],
			XY{
				X: float64(i * 3), // data is snapshot every 3s
				Y: adcToPressure(evt.Pressure.Value),
			},
		)
		raws["hygro"] = append(raws["hygro"],
			XY{
				X: float64(i * 3), // data is snapshot every 3s
				Y: adcToHygrometry(evt.Hygrometry.Raw),
			},
		)
		vals["hygro"] = append(vals["hygro"],
			XY{
				X: float64(i * 3), // data is snapshot every 3s
				Y: adcToHygrometry(evt.Hygrometry.Value),
			},
		)
		i++
	}

	for _, k := range keys {
		p, err := plot.New()
		if err != nil {
			panic(err)
		}

		p.Title.Text = k // "Temperatures"
		p.X.Label.Text = "Time (s)"
		p.Y.Label.Text = k // "Temperature (C)"
		p.Legend.Top = true

		p.Add(plotter.NewGrid())

		err = plotutil.AddLinePoints(p,
			//	"0x6404 (raw)", raws,
			"0x6401 (val)", vals[k],
		)
		if err != nil {
			log.Fatalf("error adding line-points: %v\n", err)
		}

		// Save the plot to a PNG file.
		if err := p.Save(14*vg.Inch, 8*vg.Inch, "data-"+k+".png"); err != nil {
			log.Fatalf("error saving plot: %v\n", err)
		}
	}
}

type Event struct {
	Temp       Data `json:"temp"`
	Pressure   Data `json:"pressure"`
	Hygrometry Data `json:"hygrometry"`
}

type Data struct {
	Acc    uint8 `json:"acc"`
	Avg    uint8 `json:"avg"`
	Offset int32 `json:"offset"`
	Gain   int32 `json:"gain"` // FIXME: doc says int16
	Raw    int16 `json:"raw"`
	Value  int16 `json:"value"`
}

// UnmarshalJSON decodes a JSON representation of Data.
// The official JSON format does not support hexadecimal literals.
func (d *Data) UnmarshalJSON(data []byte) error {
	r := bytes.NewReader(data[1 : len(data)-1])
	_, err := fmt.Fscanf(
		r,
		"0x%x 0x%x 0x%x 0x%x 0x%x 0x%x",
		&d.Acc, &d.Avg, &d.Offset, &d.Gain, &d.Raw, &d.Value,
	)
	return err
}

type XY struct {
	X float64
	Y float64
}

// adcToTemperature returns the temperature corresponding to a given ADC count.
//  ADC: [0; 0xFFFF) -> -10.24V;10.24V -> -20C; 80C;
func adcToTemperature(adc int16) float64 {
	return float64(adc)*0.3125e-3*10.0 - 20.0
}

// adcToPressure returns the pressure corresponding to a given ADC count.
//  ADC: [0; 0xFFFF) -> -10.24V;10.24V -> 600mbar; 1100mbar;
func adcToPressure(adc int16) float64 {
	return float64(adc)*0.3125e-3*50.0 + 600.0
}

// adcToHygrometry returns the hygrometry corresponding to a given ADC count.
//  ADC: [0; 0xFFFF) -> -10.24V;10.24V -> 0%; 100%;
func adcToHygrometry(adc int16) float64 {
	return float64(adc) * 0.3125e-3 * 10.0
}
