package model

import (
	"github.com/jmoiron/sqlx"
)

// Event represents an entire event, ie "Masters Men Age 30-39"
type Event struct {
	ID       int
	Start    string
	Name     string
	Distance int
	Bank     string
}

// Entries returns all entries for the specified event
func (event Event) Entries(db *sqlx.DB) (entries []Entry, err error) {
	err = db.Select(&entries, "SELECT * FROM entries WHERE event_id=$1", event.ID)
	return
}
