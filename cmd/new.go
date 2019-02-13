// Copyright © 2019 CJRC, Inc <greg@jrc.us>
//

package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var dbConn string

const createEntriesTableSQL = `
	CREATE TABLE IF NOT EXISTS Entries (
		id INTEGER PRIMARY KEY
	);`

const createResultsTableSQL = `
	CREATE TABLE IF NOT EXISTS Results (
		id INTEGER PRIMARY KEY
		place INTEGER DEFAULT 0
		time INTEGER DEFAULT 0
		avg_pace INTEGER DEFAULT 0
		distance INTEGER DEFAULT 0
		name VARCHAR(80) DEFAULT ""
		bib_num INTEGER UNIQUE
		class VARCHAR(20) DEFAULT ""
	);`

const createEventsTableSQL = `
CREATE TABLE IF NOT EXISTS Events (
	id INTEGER PRIMARY KEY

);`

const createRacesTableSQL = `
CREATE TABLE IF NOT EXISTS Races (
	id INTEGER PRIMARY KEY

);`

// initCmd represents the init command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Creates a new regatta (config file, database tables, etc..)",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

		// Command line flag will take precendence over env var
		if dbConn != "" {
			C.DB = dbConn
		}
		if C.DB == "" {
			fmt.Println("Must specify a database connection to create a new race.")
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

func init() {
	rootCmd.AddCommand(newCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")
	newCmd.PersistentFlags().StringVar(&dbConn, "db", "", "Connection string for the postgres database")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
