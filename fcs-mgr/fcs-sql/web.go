package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Hi there, I love %s!\n", r.URL.Path[1:])
	//fmt.Fprintf(w, "Here is my db handle: %#v\n", page.db)

	page.mu.Lock()
	defer page.mu.Unlock()

	err := page.load()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	err = page.write(w)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func startServer() error {
	http.HandleFunc("/", handler)
	const addr = "localhost:8080"
	log.Printf("starting server on [%s]...\n", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Printf("error server: %v\n", err)
		return err
	}
	return err
}

//type Point [2]float64

type Point struct {
	X time.Time
	Y float64
}

type Points []Point

func (p Points) Len() int { return len(p) }

func (p Points) Less(i, j int) bool { return p[i].X.Unix() < p[j].X.Unix() }

func (p Points) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

type Page struct {
	Title string
	db    *sql.DB
	Tmpl  *template.Template
	mu    sync.RWMutex
	id    int64 // last id

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

	stmt, err := p.db.Prepare("select * from rawdata where id > ? order by (id and descr_id)")
	if err != nil {
		return err
	}

	rows, err := stmt.Query(p.id)
	if err != nil {
		return err
	}
	defer rows.Close()

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

		switch data.DescrID {
		case 68:
			p.Temperature = append(
				p.Temperature,
				Point{
					X: data.TStamp.Time,
					Y: data.Float64.Float64,
				},
			)

		case 69:
			p.Hygrometry = append(
				p.Hygrometry,
				Point{
					X: data.TStamp.Time,
					Y: data.Float64.Float64,
				},
			)

		case 70:
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

	return err
}

func (p *Page) write(w io.Writer) error {
	return p.Tmpl.Execute(w, p)
}
