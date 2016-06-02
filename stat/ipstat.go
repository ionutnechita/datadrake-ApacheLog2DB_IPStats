package stat

import (
	"database/sql"
	"errors"
	"github.com/DataDrake/ApacheLog2DB/source"
	"github.com/DataDrake/ipstat/data"
	"github.com/DataDrake/ipstat/lms"
)

type IPStat struct {
	ID        int
	Bandwidth float64
	Latency   float64
	SourceID  int
}

func NewIPStat(bw float64, lat float64, sourceid int) *IPStat {
	return &IPStat{-1, bw, lat, sourceid}
}

func GetStats(s *source.Source) (*IPStat, error) {
	samples, err := data.CollectDataPoints(s.IP, 100, 1500, 100)
	if err != nil {
		return nil, err
	}
	stat := &IPStat{}
	slope, intercept := lms.LMS_Perf(samples)
	stat.Bandwidth = float64(1.0) / slope
	stat.Latency = intercept
	stat.SourceID = s.ID
	return stat, nil
}

func ReadOrCreate(db *sql.DB, s *source.Source) (*IPStat, error) {
	stat, err := ReadSource(db, s.ID)
	if err != nil {
		stat, err = GetStats(s)
		if err != nil {
			return nil, err
		}
		err = Insert(db, stat)
		if err == nil {
			stat, err = ReadSource(db, s.ID)
		}
	}
	return stat, err
}

func CreateTable(d *sql.DB) error {
	_, err := d.Exec("CREATE TABLE ipstats ( id INTEGER PRIMARY KEY AUTOINCREMENT," +
		"bandwidth DOUBLE, latency DOUBLE, sourceid INTEGER," +
		"FOREIGN KEY(sourceid) REFERENCES sources(id)")
	return err
}

func Insert(d *sql.DB, s *IPStat) error {
	_, err := d.Exec("INSERT INTO ipstats VALUES( NULL , ? , ? , ? , ? )", s.Bandwidth, s.Latency, s.SourceID)
	return err
}

func ReadSource(d *sql.DB, sourceid int) (*IPStat, error) {
	s := &IPStat{}
	var err error
	row := d.QueryRow("SELECT * FROM ipstats WHERE sourceid=?", sourceid)
	if row == nil {
		err = errors.New("Stats not found")
	} else {
		err = row.Scan(&s.ID, &s.Bandwidth, &s.Latency, &s.SourceID)
	}
	return s, err
}

func Read(d *sql.DB, id int) (*IPStat, error) {
	s := &IPStat{}
	var err error
	row := d.QueryRow("SELECT * FROM ipstats WHERE id=?", id)
	if row == nil {
		err = errors.New("Agent not found")
	} else {
		err = row.Scan(&s.ID, &s.Bandwidth, &s.Latency, &s.SourceID)
	}
	return s, err
}

func ReadAll(d *sql.DB) ([]*IPStat, error) {
	ss := make([]*IPStat, 0)
	rows, err := d.Query("SELECT * FROM user_agents")
	if err == nil {
		for rows.Next() {
			s := &IPStat{}
			err = rows.Scan(&s.ID, &s.Bandwidth, &s.Latency, &s.SourceID)
			if err == nil && s.ID >= 0 && s.SourceID > 0 {
				ss = append(ss, s)
			}
		}
		rows.Close()
	}
	return ss, err
}

func Update(d *sql.DB, s *IPStat) error {
	_, err := d.Exec("UPDATE ipstats SET bandwidth=? latency=? sourceid=? WHERE id=?", s.Bandwidth, s.Latency, s.SourceID, s.ID)
	return err
}