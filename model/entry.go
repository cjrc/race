package model

import (
	"sort"
	"time"

	"github.com/jmoiron/sqlx"
)

// EntrySchema is the sql commands to create the Entries table
var EntrySchema = []string{
	`CREATE TABLE Entries (
			id SERIAL PRIMARY KEY,
			email TEXT DEFAULT '',
			club_name TEXT DEFAULT '',
			club_abbrev TEXT DEFAULT '',
			seed BIGINT DEFAULT 0,
			age INTEGER DEFAULT 0,
			boat_name TEXT DEFAULT ' ',
			country TEXT DEFAULT 'USA', 
			event_id INTEGER DEFAULT 0,
			race_id INTEGER DEFAULT 0,
			lane INTEGER DEFAULT 0,
			scratched BOOLEAN DEFAULT false,
			ltwt BOOLEAN DEFAULT false,
			bib_num INTEGER UNIQUE
		);`,
	"CREATE INDEX ON Entries (race_id);",
	"CREATE INDEX ON Entries (event_id);",
	"CREATE INDEX ON Entries (bib_num);",
	// `CREATE OR REPLACE FUNCTION notify_entries() RETURNS TRIGGER AS $$
	//  BEGIN
	//    NOTIFY entries;
	//    RETURN null;
	//  END;
	//  $$ language plpgsql;`,
	// `CREATE TRIGGER notify_entries_event
	//  AFTER INSERT OR UPDATE OR DELETE ON entries
	//  FOR EACH STATEMENT EXECUTE PROCEDURE notify_entries();`,
}

// Entry represents one boat in the regatta
type Entry struct {
	ID         int           `db:"id"`
	Email      string        `db:"email"`
	ClubName   string        `db:"club_name"`
	ClubAbbrev string        `db:"club_abbrev"`
	Seed       time.Duration `db:"seed"`
	Age        int           `db:"age"`
	BoatName   string        `db:"boat_name"`
	Country    string        `db:"country"`
	EventID    int           `db:"event_id"`
	RaceID     int           `db:"race_id"`
	Lane       int           `db:"lane"`
	Scratched  bool          `db:"scratched"`
	Ltwt       bool          `db:"ltwt"`
	BibNum     int           `db:"bib_num"`
	Result     Result        `db:"result"`
}

// Insert will insert the entry into the specfied database
// Ignores conflict if bibnum already exists
// Returns true if entry was inserted
func (entry Entry) Insert(db *sqlx.DB) (bool, error) {
	sql := `INSERT INTO Entries(email, club_name, club_abbrev, seed, age, boat_name, 
		country, event_id, bib_num)
		VALUES(:email, :club_name, :club_abbrev, :seed, :age, :boat_name,
		:country, :event_id, :bib_num)
		ON CONFLICT (bib_num)
		DO NOTHING;`

	res, err := db.NamedExec(sql, &entry)
	if err != nil {
		return false, err
	}
	num, _ := res.RowsAffected()

	return (num == 1), nil
}

// SortEntriesByTime sorts the slice of entries, with the fastest finishing times coming first
// A Finish time of 0 indicates that the entry did not race, and they will be sorted to the
// end
func SortEntriesByTime(entries []Entry) {
	// Sort finishing times
	// A finishing time of 0 means they didn't start
	sort.Slice(entries, func(h, k int) bool {
		if entries[h].Result.Time == 0 {
			return false
		} else if entries[k].Result.Time == 0 {
			return true
		}
		return entries[h].Result.Time < entries[k].Result.Time
	})
}

// AssignPlacesToEntries will sort the entries by their finishing times,
// and assign places, taking ties into account
func AssignPlacesToEntries(entries []Entry) {
	SortEntriesByTime(entries)

	place := 1
	for j := range entries {
		if j == 0 {
			entries[j].Result.Place = place
			// deal with ties appropriately
		} else if entries[j].Result.Time == entries[j-1].Result.Time {
			entries[j].Result.Place = entries[j-1].Result.Place
		} else {
			entries[j].Result.Place = place
		}
		place++
	}
}
