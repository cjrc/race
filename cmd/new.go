// Copyright © 2019 CJRC, Inc <greg@jrc.us>
//

package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // database driver for Postgres
	"github.com/spf13/cobra"
)

// SQLcommands are for creating the data tables, indices, etc
var schema = []string{
	`CREATE TABLE Entries (
		id SERIAL PRIMARY KEY,
		email TEXT DEFAULT '',
		club_name TEXT DEFAULT '',
		club_abbrev TEXT DEFAULT '',
		seed TEXT DEFAULT '0:00',
		age INTEGER DEFAULT 0,
		boat_name TEXT DEFAULT ' ',
		country TEXT DEFAULT 'USA', 
		event_id INTEGER DEFAULT 0,
		race_id INTEGER DEFAULT 0,
		lane INTEGER DEFAULT 0,
		scratched BOOLEAN DEFAULT false,
		ltwt BOOLEAN DEFAULT false
	);`,
	"CREATE INDEX ON Entries (race_id);",
	"CREATE INDEX ON Entries (event_id);",
	`CREATE TABLE Results (
		id SERIAL PRIMARY KEY,
		place INTEGER DEFAULT 0,
		time INTEGER DEFAULT 0,
		avg_pace INTEGER DEFAULT 0,
		distance INTEGER DEFAULT 0,
		name text DEFAULT ''::text,
		entry_id INTEGER UNIQUE,
		class VARCHAR(20) DEFAULT ''::text
	);`,
	"CREATE INDEX ON Results (entry_id);",
	`CREATE TABLE Events (
		id SERIAL PRIMARY KEY,
		name TEXT DEFAULT ''::text,
		start TEXT DEFAULT '8:00AM'::text,
		distance INTEGER DEFAULT 2000,
		official bool DEFAULT false
	);`,
	`CREATE TABLE Races (
		id SERIAL PRIMARY KEY,
		boat_type INTEGER DEFAULT 0,
		name TEXT DEFAULT ''::text,
		distance INTEGER DEFAULT 2000,
		enable_stroke_data BOOLEAN DEFAULT false,
		split_distance INTEGER DEFAULT 500,
		split_times INTEGER DEFAULT 120,
		nlanes INTEGER DEFAULT 10,
		duration_type INTEGER DEFAULT 0
	);`,
}

// initCmd represents the init command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Creates a new regatta (config file, database tables, etc..)",
	Long: `The 'new' command will create a new regatta in the current directory.

Note:
 - The current directory must be empty.  
 - A database must be specified by either the RACE_DB environment variable or --db flag.  
 
The new command will create all the necessary tables and indices in the race database.`,
	Run: func(cmd *cobra.Command, args []string) {
		// A New regatta will be created in the current working directory
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Can't determine current directory: %s\n", err)
			os.Exit(1)
		}

		// Check to see if the directory is empty
		empty, err := isDirEmpty(pwd)
		if err != nil {
			fmt.Printf("Can't determine if current directory is empty: %s\n", err)
			os.Exit(1)
		}

		// refuse to create a new regatta in a non-empty directory
		if !empty {
			fmt.Println("Cannot create a new regatta in a non-empty directory.")
			os.Exit(1)
		}

		// we must know what database to use
		if C.DB == "" {
			fmt.Println("Must specify a database connection to create a new race.")
			os.Exit(1)
		}

		fmt.Println("creating database schema")
		if err := createDatabase(); err != nil {
			fmt.Println("Cannot create database tables:", err)
			os.Exit(1)
		}

		// Create the new regatta
		fmt.Println("creating default config")
		C.WriteToFile("race.yaml")

		fmt.Println("creating shared folders")
		if err := createSharedFolders(); err != nil {
			fmt.Printf("Can't create shared folders: %s\n", err)
		}
	},
}

func createSharedFolders() error {
	if err := os.MkdirAll(C.HTMLPath, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(C.RacePath, 0755); err != nil {
		return err
	}
	return os.MkdirAll(C.ResultsPath, 0755)
}

func isDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

func createDatabase() error {
	db, err := sqlx.Connect("postgres", C.DB)
	if err != nil {
		return err
	}

	for _, s := range schema {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(newCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
