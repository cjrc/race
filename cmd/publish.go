// Copyright © 2019 CJRC, Inc <greg@jrc.us>
//

package cmd

import (
	"fmt"
	"html/template"
	"os"
	"path"
	"sort"
	"strconv"
	"time"

	"github.com/cjrc/race/model"
	"github.com/lib/pq"
	"github.com/spf13/cobra"
)

var publishLive bool

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish shared resources (results, race files, schedule)",
	Long:  ``,
}

// publishCmd represents the publish command
var publishResultsCmd = &cobra.Command{
	Use:   "results",
	Short: "Publish regatta results as an HTML file",
	Long:  `Publishing process uses the results.html template to format results.`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if publishLive {
			err = PublishLiveResults()
		} else {
			err = PublishResults()
		}
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

// publishCmd represents the publish command
var publishScheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Publish the regatta schedule as an HTML file",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("publish called")
	},
}

// publishCmd represents the publish command
var publishRacesCmd = &cobra.Command{
	Use:   "races",
	Short: "Publish the .RAC files for Venue Racing Application",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("publish called")
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)
	publishCmd.AddCommand(publishResultsCmd)
	publishCmd.AddCommand(publishRacesCmd)
	publishCmd.AddCommand(publishScheduleCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// publishCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// publishCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	publishResultsCmd.Flags().BoolVar(&publishLive, "live", false, "Publish live results in realtime.")
}

func durString(d time.Duration) string {
	if d == 9999*time.Hour { // this represents someone that didn't start
		return "-"
	}
	mins := (d / time.Minute)
	secs := (d - mins*time.Minute).Seconds()
	return fmt.Sprintf("%d:%04.1f", mins, secs)
}

// PublishResults creates a nice HTML view of the results in the folder specified by path
// TODO: Imported from indoor-2019, fix it up
func PublishResults() error {

	// Events sorted by their event number
	var events = append([]model.Event(nil), C.Events...)

	sort.Slice(events, func(h, k int) bool {
		return events[h].ID < events[k].ID
	})

	db := DBMustConnect()

	for i := range events {
		// Load the entries for this event
		if err := events[i].LoadEntriesWithResults(db); err != nil {
			return err
		}

		// Sort finishing times
		// A finishing time of 0 means they didn't start
		sort.Slice(events[i].Entries, func(h, k int) bool {
			if events[i].Entries[h].Result.Time == 0 {
				return false
			} else if events[i].Entries[k].Result.Time == 0 {
				return true
			}
			return events[i].Entries[h].Result.Time < events[i].Entries[k].Result.Time
		})

		// Give each result a finish place
		place := 1
		for j := range events[i].Entries {
			if j == 0 {
				events[i].Entries[j].Result.Place = place
				// deal with ties appropriately
			} else if events[i].Entries[j].Result.Time == events[i].Entries[j-1].Result.Time {
				events[i].Entries[j].Result.Place = events[i].Entries[j-1].Result.Place
			} else {
				events[i].Entries[j].Result.Place = place
			}
			place++
		}
	}

	// create the HTML results file
	fullname := path.Join(C.HTMLPath, "results.html")
	fmt.Println("Publishing results to", fullname)

	file, err := os.Create(fullname)
	if err != nil {
		return err
	}
	defer file.Close()

	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
		"needBreak": func(i int) bool {
			return (i % 2) == 0
		},
		"official": func(e model.Event) template.HTML {
			//TODO FIX THIS
			// if e.Official {
			// 	return template.HTML("Official")
			// }
			return template.HTML("<i>Unofficial</i>")
		},
		"place": func(e model.Entry) string {
			if e.Result.Time == 0 {
				return "-"
			}
			return strconv.Itoa(e.Result.Place)
		},
		"now": func() string {
			return time.Now().Format("Jan 2, 2006 at 03:04PM")
		},
		"ltwt": func(entry model.Entry) string {
			if entry.Ltwt {
				return "(Ltwt)"
			}
			return ""
		},
		"time": func(entry model.Entry) string {
			if entry.Result.Time == 0 {
				return "-"
			}
			return durString(entry.Result.Time)
		},
	}

	templatePath := path.Join(C.TemplatePath, "results.html")
	t, err := template.New(path.Base(templatePath)).Funcs(funcMap).ParseFiles(templatePath)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	data["Events"] = events
	err = t.Execute(file, data)

	return err

}

func waitForResults(l *pq.Listener) error {
	for {
		select {
		case <-l.Notify:
			if err := PublishResults(); err != nil {
				return err
			}
		case <-time.After(5 * time.Minute):
			fmt.Println("Received no results for 5 minutes, checking connection")
			if err := l.Ping(); err != nil {
				return err
			}
		}
	}
}

// PublishLiveResults will publish HTML results as they arrive at the database
func PublishLiveResults() error {
	// publish existing results
	if err := PublishResults(); err != nil {
		return err
	}

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	listener := pq.NewListener(C.DB, 10*time.Second, time.Minute, reportProblem)
	if err := listener.Listen("results"); err != nil {
		return err
	}

	fmt.Println("Listening for live results...")
	return waitForResults(listener)

}
