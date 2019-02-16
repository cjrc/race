package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cjrc/race/model"
	"github.com/extrame/xls"
	"github.com/spf13/cobra"
)

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

func addEntriesToDatabase(entries []model.Entry) error {
	db := DBMustConnect()

	for _, entry := range entries {
		// ignore empty results
		if entry.BibNum == 0 {
			continue
		}
		fmt.Printf("Adding Entry for %s (bib # %d)..", entry.BoatName, entry.BibNum)
		ok, err := entry.Insert(db)
		if err != nil {
			return err
		}

		if !ok {
			fmt.Println(" duplicate entry, ignored.")
		} else {
			fmt.Println(" done.")
		}
	}

	return nil

}

func importRows(rows [][]string) error {
	var entries []model.Entry

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

		entry := model.Entry{
			EventID:    eventID,
			Email:      row[C.EntryCols.Email],
			ClubName:   row[C.EntryCols.ClubName],
			ClubAbbrev: row[C.EntryCols.ClubAbbrev],
			Seed:       seed,
			Age:        age,
			BoatName:   row[C.EntryCols.BoatName],
			Country:    row[C.EntryCols.Country],
			BibNum:     boatID,
		}

		entries = append(entries, entry)
	}
	return addEntriesToDatabase(entries)
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

	importCmd.PersistentFlags().StringVar(&EntriesFilename, "file", EntriesFilename, "Path to Excel file from Regatta Central")
}
