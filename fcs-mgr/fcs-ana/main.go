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

type Data struct {
	Acc    uint8 `json:"acc"`
	Avg    uint8 `json:"avg"`
	Offset int32 `json:"offset"`
	Gain   int32 `json:"gain"` // FIXME: doc says int16
	Raw    int16 `json:"raw"`
	Value  int16 `json:"value"`
}

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

func main() {
	fname := os.Args[1]
	f, err := os.Open(fname)
	if err != nil {
		log.Fatalf("could not open [%s]: %v\n", fname, err)
	}
	defer f.Close()

	raws := make(plotter.XYs, 0, 1024)
	vals := make(plotter.XYs, 0, 1024)

	i := 0
	scan := bufio.NewScanner(f)
	for scan.Scan() {
		line := scan.Bytes()
		if bytes.HasPrefix(line, []byte("#")) {
			continue
		}
		/*
			// FIXME(sbinet): stdout and stderr are mixed up in fcs-mgr...
			if !bytes.HasPrefix(line, []byte("0x3 0x4 0x5b0000 0xfdae")) {
				continue
			}
		*/
		r := bytes.NewReader(line)
		data := make(map[string]Data)
		err = json.NewDecoder(r).Decode(&data)
		if err != nil {
			log.Fatalf("error decoding line %q: %v\n", string(line), err)
		}
		/*
			_, err = fmt.Fscanf(
				r,
				"0x%x 0x%x 0x%x 0x%x 0x%x 0x%x",
				&data.Acc, &data.Avg, &data.Offset, &data.Gain,
				&data.Raw, &data.Value,
			)
			if err != nil {
				log.Printf("error while scanning line %q: %v\n", string(line), err)
				continue
			}
		*/
		raws = append(raws,
			XY{
				X: float64(i),
				Y: float64(data["temp"].Raw)*0.3125e-3*10 - 20,
			},
		)
		vals = append(vals,
			XY{
				X: float64(i),
				Y: float64(data["temp"].Value)*0.3125e-3*10 - 20,
			},
		)
		i++
	}

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Temperatures"
	p.X.Label.Text = "time"
	p.Y.Label.Text = "Temperatures"

	err = plotutil.AddLinePoints(p,
		"0x6404 (raw)", raws,
		"0x6401 (val)", vals,
	)
	if err != nil {
		log.Fatalf("error adding line-points: %v\n", err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(14*vg.Inch, 8*vg.Inch, "data.png"); err != nil {
		log.Fatalf("error saving plot: %v\n", err)
	}

}
