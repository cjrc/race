// Copyright © 2019 CJRC, Inc <greg@jrc.us>
//

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cjrc/erg"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var doneCh chan bool
var liveResults bool

// tallyCmd represents the tally command
var resultsCmd = &cobra.Command{
	Use:   "results",
	Short: "Import the race results for each event",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := importResults(); err != nil {
			fmt.Println("Error importing results:", err)
			os.Exit(1)
		}
		if liveResults {
			importLiveResults()
		}
	},
}

func addResultsToDatabase(results []erg.Result) error {
	// TODO: add sql code here
	return nil
}

func importResults() error {
	filenames, err := filepath.Glob(filepath.Join(C.ResultsPath, "*.txt"))
	if err != nil {
		return err
	}

	for _, filename := range filenames {
		results, err := erg.ReadResultsFromFile(filename)
		if err != nil {
			return err
		}
		if err := addResultsToDatabase(results); err != nil {
			return err
		}
	}

	return nil
}

func watchResults(watcher *fsnotify.Watcher) {
	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				results, err := erg.ReadResultsFromFile(event.Name)
				if err != nil {
					fmt.Println("Error reading results:", err)
					doneCh <- true
				}
				if err := addResultsToDatabase(results); err != nil {
					fmt.Println("Error saving results:", err)
					doneCh <- true
				}
			}
		case err := <-watcher.Errors:
			fmt.Println("Error watching:", err)
			doneCh <- true
		}
	}
}

func importLiveResults() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Cannot create watcher:", err)
		os.Exit(1)
	}
	defer watcher.Close()

	doneCh = make(chan bool)
	go watchResults(watcher)

	err = watcher.Add(C.ResultsPath)
	if err != nil {
		fmt.Println("Cannot watch results:", err)
		os.Exit(1)
	}

	<-doneCh
}

func init() {
	importCmd.AddCommand(resultsCmd)

	resultsCmd.Flags().BoolVar(&liveResults, "live", false, "Watch the results path and tally events as new results arrive")

}
