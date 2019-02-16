package model

import (
	"github.com/cjrc/erg"
	"github.com/jmoiron/sqlx"
)

// Result is an erg race race results as reported by the Venue software
type Result struct {
	erg.Result
}

// Insert will insert the result into the specfied database
// Ignores conflict if bibnum already exists
// Returns true if result was inserted
func (result Result) Insert(db *sqlx.DB) (bool, error) {
	sql := `INSERT INTO Results(place, time, avg_pace, distance, name, bib_num, class) 
			VALUES(:place, :time, :avg_pace, :distance, :name, :bib_num, :class)
			ON CONFLICT (bib_num)
			DO NOTHING;`

	res, err := db.NamedExec(sql, &result)
	if err != nil {
		return false, err
	}
	num, _ := res.RowsAffected()

	return (num == 1), nil
}
