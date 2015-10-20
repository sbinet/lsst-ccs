// fcs-parse-msr-sensor parses CSV files recorded from the MSR sensor
package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	fname = flag.String("f", "MSR453196_150729_150928.csv", "path to MSR sensor data file")
)

type Column struct {
	Name      string // title of the associated data
	Unit      string // units of the associated data
	Sensor    string // name of the sensor collecting the data
	SensorID  string // id of the sensor collecting the data
	TimeDelay time.Duration
	Limits    Limits
	CalibData CalibData
}

type Row struct {
	Time time.Time
	Data []float64
}

type Limits struct {
	Alarm    float64
	Recorded float64
	Limit1   float64
	Limit2   float64
}

type CalibData struct {
	Info string
	Date time.Time
	X0   float64
	Y0   float64
	X1   float64
	Y1   float64
}

func main() {
	flag.Parse()

	f, err := os.Open(*fname)
	if err != nil {
		log.Fatalf("error opening MSR data file [%s]: %v\n", *fname, err)
	}
	defer f.Close()

	var cols []Column
	var rows []Row

	var section Section
	scan := bufio.NewScanner(f)
	for scan.Scan() {
		line := scan.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if line[0] == '*' {
			switch line {
			case "*CREATOR":
				log.Printf(">>> creator section\n")
				section = CreatorSection

			case "*STARTTIME":
				log.Printf(">>> start-time section\n")
				section = StartTimeSection

			case "*MODUL":
				log.Printf(">>> module section\n")
				section = ModuleSection

			case "*NAME":
				log.Printf(">>> name section\n")
				section = NameSection

			case "*TIMEDELAY":
				log.Printf(">>> time-delay section\n")
				section = TimeDelaySection

			case "*CHANNEL":
				log.Printf(">>> channel section\n")
				section = ChannelSection

			case "*UNIT":
				log.Printf(">>> unit section\n")
				section = UnitSection

			case "*LIMITS":
				log.Printf(">>> limits section\n")
				section = LimitsSection

			case "*CALIBRATION":
				log.Printf(">>> calibration section\n")
				section = CalibrationSection

			case "*DATA":
				log.Printf(">>> data section\n")
				section = DataSection
			}
			continue
		}

		tokens := strings.Split(line, ";")

		switch section {
		case CreatorSection:
		case StartTimeSection:
			start, err := time.Parse("2006-01-02;15:04:05;", line)
			if err != nil {
				log.Fatalf("error parsing start-time: %v\nline=%q\n", err, line)
			}
			log.Printf("start-time: %v\n", start)

		case ModuleSection:
			cols = make([]Column, len(tokens))
			for i, tok := range tokens {
				log.Printf("    module: %q\n", tok)
				cols[i].Sensor = tok
			}

		case NameSection:

		case TimeDelaySection:
			for i, tok := range tokens {
				if i > 0 {
					delay, err := time.ParseDuration(tok + tokens[0])
					if err != nil {
						log.Fatalf(
							"error parsing time-delay: %v\nline=%q\n",
							err, line,
						)
					}
					cols[i].TimeDelay = delay
				}
			}

		case ChannelSection:
			for i, tok := range tokens {
				log.Printf("    channel: %q\n", tok)
				cols[i].Name = tok
			}

		case UnitSection:
			for i, tok := range tokens {
				if i == 0 {
					continue
				}
				cols[i].Unit = tok
			}

		case LimitsSection:
			// TODO(sbinet)
		case CalibrationSection:
			// TODO(sbinet)
		case DataSection:
			var row Row
			row.Time, err = time.Parse("2006-01-02 15:04:05.999", tokens[0])
			if err != nil {
				log.Fatalf("error parsing data: %v\nline=%q\n", err, line)
			}
			row.Data = make([]float64, len(tokens)-1)
			for i, tok := range tokens[1:] {
				switch tok {
				case "":
					row.Data[i] = getLastOrDefault(rows, i)
				default:
					val, err := strconv.ParseFloat(tok, 64)
					if err != nil {
						log.Fatalf(
							"error parsing float data: %v\nline=%q\n",
							err,
							line,
						)
					}
					row.Data[i] = val
				}
			}
			log.Printf("data: %v %v\n", row.Time, row.Data)
			rows = append(rows, row)
		}

	}
	err = scan.Err()
	if err == io.EOF {
		err = nil
	}
	if err != nil {
		log.Fatalf("error scanning MSR data file [%s]: %v\n", *fname, err)
	}

	for i, col := range cols {
		log.Printf("col[%d] = %#v\n", i, col)
	}
}

type Section int

const (
	UndefinedSection Section = iota
	CreatorSection
	StartTimeSection
	ModuleSection
	NameSection
	TimeDelaySection
	ChannelSection
	UnitSection
	LimitsSection
	CalibrationSection
	DataSection
)

func getLastOrDefault(rows []Row, icol int) float64 {
	if len(rows) == 0 {
		return 0.0
	}
	return rows[len(rows)-1].Data[icol]
}
