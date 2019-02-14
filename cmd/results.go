// Copyright © 2019 CJRC, Inc <greg@jrc.us>
//

package cmd

import (
	"fmt"
	"log"
	"os"

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
	},
}

func watchResults(watcher *fsnotify.Watcher) {
	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				// process the results file here
				log.Println("modified file:", event.Name)
			}
		case err := <-watcher.Errors:
			fmt.Println("Error watching:", err)
			doneCh <- true
		}
	}
}

func init() {
	importCmd.AddCommand(resultsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tallyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tallyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	resultsCmd.Flags().BoolVar(&liveResults, "live", false, "Watch the results path and tally events as new results arrive")

}
