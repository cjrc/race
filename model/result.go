package model

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// ResultSchema is the sql commands to create the Results table
var ResultSchema = []string{
	`CREATE TABLE Results (
		id SERIAL PRIMARY KEY,
		place INTEGER DEFAULT 0,
		time BIGINT DEFAULT 0,
		avg_pace BIGINT DEFAULT 0,
		distance INTEGER DEFAULT 0,
		name text DEFAULT ''::text,
		bib_num INTEGER UNIQUE,
		class VARCHAR(20) DEFAULT ''::text,
		official BOOLEAN DEFAULT false,
	);`,
	// "CREATE INDEX ON Results (bib_num);",
	// `CREATE OR REPLACE FUNCTION notify_results() RETURNS TRIGGER AS $$
	//  BEGIN
	//    NOTIFY results;
	//    RETURN null;
	//  END;
	//  $$ language plpgsql;`,
	// `CREATE TRIGGER notify_results_event
	//  AFTER INSERT OR UPDATE OR DELETE ON results
	//  FOR EACH STATEMENT EXECUTE PROCEDURE notify_results();`,
}

// Result represents the race result of one erg from the Venue Racing results file
type Result struct {
	Place    int           `db:"place"`
	Time     time.Duration `db:"time"`
	AvgPace  time.Duration `db:"avg_pace"`
	Distance int           `db:"distance"`
	Name     string        `db:"name"`
	BibNum   int           `db:"bib_num"`
	Class    string        `db:"class"`

	// Not used by the Venue racing app
	Official *bool `db:"official"`
}

//ReadResults reads the race results from the specified io.Reader and appends them to the
//supplied results array.
//It will return an error if the read results are in an invalid format or are not
//Version 103 results.
func ReadResults(results *[]Result, reader io.Reader) error {
	scanner := bufio.NewScanner(reader)

	scanner.Scan()
	if scanner.Text() != "Race Results" {
		return fmt.Errorf("invalid or corrupted race results")
	}

	scanner.Scan()
	if ver := scanner.Text(); ver != "103" {
		return fmt.Errorf("found version %v results -- this software only knows version 103", ver)
	}

	scanner.Scan() // Skip blank line
	scanner.Scan() // Skip headers line

	lineNumber := 5
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break // a blank line is the end of results
		}

		parts := strings.Split(line, ",")
		if len(parts) != 8 {
			return fmt.Errorf("invalid results on line %d", lineNumber)
		}

		tmp := strings.Replace(parts[1], ":", "m", -1) + "s"
		finishTime, err := time.ParseDuration(tmp)
		if err != nil {
			return fmt.Errorf("invalid race time on line %d: %v", lineNumber, err)
		}

		// remove spaces and set the minutes marker to the format wanted by Go
		tmp = strings.TrimSpace(strings.Replace(parts[4], ":", "m", -1))
		if tmp == "" { // if there is no pace, set it to 0
			tmp = "0"
		}
		tmp += "s" // add seconds marker
		avgPace, err := time.ParseDuration(tmp)
		if err != nil {
			return fmt.Errorf("invalid average pace on line %d: %v", lineNumber, err)
		}

		distance, err := strconv.Atoi(parts[2])
		if err != nil {
			return fmt.Errorf("invalid race distance on line %d: %v", lineNumber, err)
		}

		place, err := strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Errorf("invalid finish place on line %d: %v", lineNumber, err)
		}

		bibNum, err := strconv.Atoi(parts[6])
		if err != nil {
			return fmt.Errorf("invalid id on line %d: %v", lineNumber, err)
		}

		result := Result{
			Place:    place,
			Time:     finishTime,
			Distance: distance,
			Name:     parts[3],
			AvgPace:  avgPace,
			BibNum:   bibNum,
			Class:    parts[7],
		}
		*results = append(*results, result)

		lineNumber++
	}
	return scanner.Err()
}

// ReadResultsFromFile is a convenience function reads result from the specified filed
func ReadResultsFromFile(filename string) ([]Result, error) {
	results := make([]Result, 0)

	file, err := os.Open(filename)
	if err != nil {
		return results, err
	}

	defer file.Close()

	err = ReadResults(&results, file)

	return results, err
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

// NotifyResults will send the 'results' notification to the DB
func NotifyResults(db *sqlx.DB) error {
	_, err := db.Exec("NOTIFY results;")
	return err
}
