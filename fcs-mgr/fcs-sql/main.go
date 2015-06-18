package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var (
	user = flag.String("user", "", "db user name")
	pass = flag.String("password", "", "db user password")
)

func main() {

	flag.Parse()
	conn := *user + ":" + *pass + "@/ccs"
	db, err := sql.Open("mysql", conn)
	if err != nil {
		log.Fatalf("error opening db connection: %v\n", err)
	}
	defer db.Close()

	// Open doesn't open a connection. Validate DSN data:
	err = db.Ping()
	if err != nil {
		log.Fatalf("error pinging db: %v\n", err)
	}

	descr, err := loadDataDesc(db)
	if err != nil {
		log.Fatalf("error loading rawdata descriptions: %v\n", err)
	}

	stmt, err := db.Prepare("select * from rawdata where id > 4680000 order by id and descr_id")
	if err != nil {
		log.Fatalf("error preparing stmt: %v\n", err)
	}

	rows, err := stmt.Query()
	if err != nil {
		log.Fatalf("error in query: %v\n", err)
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
			log.Fatalf("error scanning: %v\n", err)
		}
		var v interface{}
		if data.Float64.Valid {
			v = data.Float64.Float64
		}
		if data.String.Valid {
			v = data.String.String
		}
		dataDesc := descr[data.DescrID]
		name := "???"
		if dataDesc.Name.Valid {
			name = strings.Replace(dataDesc.Name.String, "testbenchLPC/", "", -1)
		}

		fmt.Printf(
			"%d \"%v\" %-15s = %v\n",
			data.ID,
			data.TStamp.Time,
			name,
			v,
		)
	}

}

func loadDataDesc(db *sql.DB) (map[int64]DataDesc, error) {
	stmt, err := db.Prepare("select * from datadesc order by id")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	descr := make(map[int64]DataDesc)
	for rows.Next() {
		var data DataDesc
		err = rows.Scan(
			&data.ID,
			&data.Type,
			&data.MaxSampling,
			&data.Name,
			&data.PDelay,
			&data.SrcName,
			&data.SrcSubSystem,
		)
		if err != nil {
			return nil, err
		}
		descr[data.ID] = data
	}

	return descr, nil
}
