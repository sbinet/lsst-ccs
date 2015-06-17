package main

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"time"
)

// Timestamp represents a time stamp that may be null.
// Timestamp implements the sql.Scanner interface so it can be used as a Scan
// desctination.
type Timestamp struct {
	Valid bool
	Time  time.Time
}

func (ts *Timestamp) Scan(value interface{}) error {
	if value == nil {
		ts.Time, ts.Valid = time.Time{}, false
		return nil
	}
	ts.Valid = true
	rv := reflect.ValueOf(value)
	if !rv.Type().ConvertibleTo(reflect.TypeOf(int64(0))) {
		return fmt.Errorf("%T is not convertible to int64", value)
	}
	ts.Time = time.Unix(rv.Int()/1000.0, 0)
	return nil
}

func (ts Timestamp) Value() (driver.Value, error) {
	if !ts.Valid {
		return nil, nil
	}
	return ts.Time, nil
}

// +-------------------+--------------+------+-----+---------+----------------+
// | Field             | Type         | Null | Key | Default | Extra          |
// +-------------------+--------------+------+-----+---------+----------------+
// | id                | bigint(20)   | NO   | PRI | NULL    | auto_increment |
// | dataType          | varchar(1)   | YES  |     | NULL    |                |
// | maxSamplingMillis | int(11)      | YES  |     | 0       |                |
// | name              | varchar(255) | YES  | UNI | NULL    |                |
// | preservationDelay | int(11)      | YES  |     | 0       |                |
// | srcName           | varchar(255) | YES  |     | NULL    |                |
// | srcSubsystem      | varchar(255) | YES  |     | NULL    |                |
// +-------------------+--------------+------+-----+---------+----------------+
type DataDesc struct {
	ID           int64
	Type         sql.NullString
	MaxSampling  time.Duration
	Name         sql.NullString
	PDelay       time.Duration
	SrcName      sql.NullString
	SrcSubSystem sql.NullString
}

// +--------------+--------------+------+-----+---------+----------------+
// | Field        | Type         | Null | Key | Default | Extra          |
// +--------------+--------------+------+-----+---------+----------------+
// | id           | bigint(20)   | NO   | PRI | NULL    | auto_increment |
// | name         | varchar(255) | YES  |     | NULL    |                |
// | tstartmillis | bigint(20)   | YES  |     | NULL    |                |
// | tstopmillis  | bigint(20)   | YES  |     | NULL    |                |
// | value        | varchar(255) | YES  |     | NULL    |                |
// | rawDescr_id  | bigint(20)   | YES  | MUL | NULL    |                |
// +--------------+--------------+------+-----+---------+----------------+
type MetaData struct {
	ID    int64
	Name  sql.NullString
	Start Timestamp
	Stop  Timestamp
	Value sql.NullString
	RawID sql.NullInt64
}

// +-------------+--------------+------+-----+---------+----------------+
// | Field       | Type         | Null | Key | Default | Extra          |
// +-------------+--------------+------+-----+---------+----------------+
// | id          | bigint(20)   | NO   | PRI | NULL    | auto_increment |
// | doubleData  | double       | YES  |     | NULL    |                |
// | stringData  | varchar(255) | YES  |     | NULL    |                |
// | tstampmills | bigint(20)   | YES  |     | NULL    |                |
// | descr_id    | bigint(20)   | YES  | MUL | NULL    |                |
// +-------------+--------------+------+-----+---------+----------------+
type RawData struct {
	ID      int64
	Float64 sql.NullFloat64
	String  sql.NullString
	TStamp  Timestamp
	DescrID int64
}

// +--------------+------------+------+-----+---------+----------------+
// | Field        | Type       | Null | Key | Default | Extra          |
// +--------------+------------+------+-----+---------+----------------+
// | id           | bigint(20) | NO   | PRI | NULL    | auto_increment |
// | data         | double     | NO   |     | NULL    |                |
// | n            | int(11)    | NO   |     | NULL    |                |
// | sum2         | double     | NO   |     | NULL    |                |
// | tstampmills1 | bigint(20) | YES  |     | NULL    |                |
// | tstampmills2 | bigint(20) | YES  |     | NULL    |                |
// | descr_id     | bigint(20) | YES  | MUL | NULL    |                |
// +--------------+------------+------+-----+---------+----------------+
type StatData struct {
	ID      int64
	Data    float64
	N       int64
	Sum2    float64
	TStamp1 Timestamp
	TStamp2 Timestamp
	DescrID int64
}

// +-------------------+------------+------+-----+---------+----------------+
// | Field             | Type       | Null | Key | Default | Extra          |
// +-------------------+------------+------+-----+---------+----------------+
// | id                | bigint(20) | NO   | PRI | NULL    | auto_increment |
// | preservationDelay | int(11)    | YES  |     | 0       |                |
// | timeBinWidth      | bigint(20) | NO   |     | NULL    |                |
// | rawDescr_id       | bigint(20) | YES  | MUL | NULL    |                |
// +-------------------+------------+------+-----+---------+----------------+
type StatDesc struct {
	ID           int64
	PDelay       time.Duration
	TimeBinWidth time.Duration
	RawID        sql.NullInt64
}
