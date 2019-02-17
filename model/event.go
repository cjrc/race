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
	Entries  []Entry `yaml:"entries,omitempty"`
}

// LoadEntriesWithResults populates the Entries field for the specified event
func (event *Event) LoadEntriesWithResults(db *sqlx.DB) (err error) {
	sql := `
SELECT 
	entries.*,
	results.place "result.place",
	results.time "result.time",
	results.avg_pace "result.avg_pace",
	results.distance "result.distance",
	results.name "result.name",
	results.bib_num "result.bib_num",
	results.class "result.class",
	results.official "result.official"
FROM
	entries JOIN results ON entries.bib_num = results.bib_num
WHERE
	event_id=$1`

	event.Entries = nil // so entries cannot be double loaded
	err = db.Select(&event.Entries, sql, event.ID)
	return
}
