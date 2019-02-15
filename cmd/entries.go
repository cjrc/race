package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/extrame/xls"
	"github.com/spf13/cobra"
)

// Entry represents one boat in the regatta
type Entry struct {
	ID         int
	Email      string
	ClubName   string
	ClubAbbrev string
	Seed       time.Duration
	Age        int
	Name       string
	Country    string
	EventID    int
	RaceID     int
	Lane       int
	Scratched  bool
	Ltwt       bool
	BibNum     int
}

// EntriesFilename is the location of the excel workbook containing entries.
// Defaults to "boats.xls"
var EntriesFilename = "boats.xls"

// importCmd represents the import command
var entriesCmd = &cobra.Command{
	Use:   "entries",
	Short: "Import race entries from RegattaCentral",
	Long: `The import entries command reads rows from the specified Excel workbook 
and saves them to the database. The expected file format is an .XLS workbook,
downloaded from RegattaCentral as "Generic Boats".`,
	Run: func(cmd *cobra.Command, args []string) {
		importEntries()
	},
}

func importRows(rows [][]string) error {
	// ignore the header row
	for rowid, row := range rows[1:] {
		ErrorRow := rowid + 2 // for error reporting, the row # as soon in Excel

		if row[C.EntryCols.EventID] == "" {
			continue //ignore empty rows
		}

		eventID, err := strconv.Atoi(row[C.EntryCols.EventID])
		if err != nil {
			return fmt.Errorf("Row %d, invalid event id: '%v'", ErrorRow, row[C.EntryCols.EventID])
		}

		boatID, err := strconv.Atoi(row[C.EntryCols.BoatID])
		if err != nil {
			return fmt.Errorf("Row %d, invalid boat id: '%v'", ErrorRow, row[C.EntryCols.BoatID])

		}

		age, err := strconv.Atoi(row[C.EntryCols.Age])
		if err != nil {
			return fmt.Errorf("Row %d, invalid age: %v", ErrorRow, row[C.EntryCols.Age])
		}

		seed, err := time.ParseDuration(strings.Replace(row[C.EntryCols.Seed], ":", "m", 1) + "s")
		if err != nil {
			return fmt.Errorf("Row %d, invalid seed time: %v", ErrorRow, row[C.EntryCols.Seed])
		}

		entry := Entry{
			EventID:    eventID,
			Email:      row[C.EntryCols.Email],
			ClubName:   row[C.EntryCols.ClubName],
			ClubAbbrev: row[C.EntryCols.ClubAbbrev],
			Seed:       seed,
			Age:        age,
			Name:       row[C.EntryCols.BoatName],
			Country:    row[C.EntryCols.Country],
			BibNum:     boatID,
		}

		fmt.Println("Importing ", entry.Name)
		// TODO create in the db
	}
	return nil
}

func importEntries() {
	workbook, err := xls.Open(EntriesFilename, "utf-8")

	if err != nil {
		fmt.Println("Cannot open XLS workbook:", err)
		os.Exit(1)
	}

	rows := workbook.ReadAllCells(C.MaxEntries)

	if err := importRows(rows); err != nil {
		fmt.Println("Error importing entries:", err)
		os.Exit(1)
	}

}

func init() {
	importCmd.AddCommand(entriesCmd)

	importCmd.PersistentFlags().StringVar(&EntriesFilename, "file", EntriesFilename, "Specify Excel workbook with entries from Regatta Central")
}
