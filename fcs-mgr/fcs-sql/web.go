package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

func handler(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Hi there, I love %s!\n", r.URL.Path[1:])
	//fmt.Fprintf(w, "Here is my db handle: %#v\n", page.db)

	page.mu.Lock()
	defer page.mu.Unlock()

	//	err := page.load()
	//	if err != nil {
	//		http.Error(w, err.Error(), 500)
	//		return
	//	}

	err := page.write(w)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func dataHandler(ws *websocket.Conn) {
	buf := new(bytes.Buffer)
	for data := range datac {
		buf.Reset()
		// fmt.Printf("Sending to client: %#v\n", data)

		err := json.NewEncoder(buf).Encode(data)
		if err != nil {
			log.Printf("error encoding data: %v\n", err)
			continue
		}

		err = websocket.Message.Send(ws, string(buf.Bytes()))
		if err != nil {
			log.Printf("Can't send: %v\n", err)
			break
		}
		//	fmt.Printf("---[temp]: %v\n", string(buf.Bytes()))
	}
}

func startServer() error {
	http.HandleFunc("/", handler)
	http.Handle("/data", websocket.Handler(dataHandler))

	const addr = "127.0.0.1:8080"
	log.Printf("starting server on [%s]...\n", addr)
	page.URI = "ws://" + addr + "/data"
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Printf("error server: %v\n", err)
		return err
	}
	return err
}

type Point struct {
	X time.Time
	Y float64
}

func (p Point) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode([2]interface{}{
		p.X.Unix() * 1000, // javascript's time expects data in milliseconds
		p.Y,
	})
	return buf.Bytes(), err
}

type Points []Point

func (p Points) Len() int { return len(p) }

func (p Points) Less(i, j int) bool { return p[i].X.Unix() < p[j].X.Unix() }

func (p Points) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

type Page struct {
	Title string
	URI   string
	db    *sql.DB
	Tmpl  *template.Template
	mu    sync.RWMutex
	id    int64              // last id
	descr map[int64]DataDesc // descr-id -> description

	Temperature Points
	Pressure    Points
	Hygrometry  Points
}

func (p *Page) load() error {
	fmt.Printf("... loading db data ... (id=%d,len=%d)\n", p.id, len(p.Temperature))
	start := time.Now()

	//p.Temperature = p.Temperature[:p.id]
	//p.Pressure = p.Pressure[:p.id]
	//p.Hygrometry = p.Hygrometry[:p.id]

	if len(p.descr) == 0 {
		descr, err := loadDataDesc(p.db)
		if err != nil {
			return err
		}
		p.descr = descr
	}

	stmt, err := p.db.Prepare("select * from rawdata where id > ? order by (id and descr_id)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	fmt.Printf("... query...\n")
	rows, err := stmt.Query(p.id)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Printf("... row-range...\n")
	for rows.Next() {
		var data RawData
		err = rows.Scan(
			&data.ID,
			&data.Float64, &data.String, &data.TStamp,
			&data.DescrID,
		)
		if err != nil {
			return err
		}

		name := ""
		descr := p.descr[data.DescrID]
		if descr.Name.Valid {
			name = descr.Name.String
		}

		switch name {
		case "testbenchLPC/temperature":
			p.Temperature = append(
				p.Temperature,
				Point{
					X: data.TStamp.Time,
					Y: data.Float64.Float64,
				},
			)

		case "testbenchLPC/hygrometry":
			p.Hygrometry = append(
				p.Hygrometry,
				Point{
					X: data.TStamp.Time,
					Y: data.Float64.Float64,
				},
			)

		case "testbenchLPC/pressure":
			p.Pressure = append(
				p.Pressure,
				Point{
					X: data.TStamp.Time,
					Y: data.Float64.Float64,
				},
			)
		}
		if p.id < data.ID {
			p.id = data.ID
		}
	}
	sort.Sort(p.Temperature)
	sort.Sort(p.Pressure)
	sort.Sort(p.Hygrometry)

	delta := time.Since(start)
	fmt.Printf("... loading db data ...[done] (%v)\n", delta)

	datac <- map[string]interface{}{
		"temperature": p.Temperature,
		"pressure":    p.Pressure,
		"hygrometry":  p.Hygrometry,
	}

	if *verbose {
		fmt.Printf("temp:  %v C\n", p.Temperature[len(p.Temperature)-1].Y)
		fmt.Printf("hygro: %v %%\n", p.Hygrometry[len(p.Hygrometry)-1].Y)
		fmt.Printf("press: %v mbar\n", p.Pressure[len(p.Pressure)-1].Y)
	}
	return err
}

func (p *Page) write(w io.Writer) error {
	return p.Tmpl.Execute(w, p)
}
