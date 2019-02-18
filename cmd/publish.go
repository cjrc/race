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

		// Sort entries and give each result a finish place
		model.AssignPlacesToEntries(events[i].Entries)
	}

	// create the HTML results file
	fullname := path.Join(C.HTMLPath, "results.html")
	file, err := os.Create(fullname)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Println("Publishing results to", fullname)

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
		fmt.Println("Listening for live results...")
		select {
		case <-l.Notify:
			if err := PublishResults(); err != nil {
				return err
			}
		case <-time.After(5 * time.Minute):
			fmt.Print("Received no results for 5 minutes, checking connection... ")
			if err := l.Ping(); err != nil {
				return err
			}
			fmt.Println("connection good!")
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
	defer listener.Close()

	// listen for changes to results and entries
	if err := listener.Listen("results"); err != nil {
		return err
	}
	if err := listener.Listen("entries"); err != nil {
		return err
	}

	return waitForResults(listener)
}
