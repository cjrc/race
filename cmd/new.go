// Copyright © 2019 CJRC, Inc <greg@jrc.us>
//

package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cjrc/race/model"
	_ "github.com/lib/pq" // database driver for Postgres
	"github.com/spf13/cobra"
)

var resultsTemplate = `
<html>
    <head>
        <title>
            2019 Cincinnati Indoor Rowing Championship Results
        </title>
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <link rel="stylesheet" href="https://www.w3schools.com/w3css/4/w3.css">
    </head>

    <body>

        <div class="w3-container">
            <h2>2019 Cincinnati Indoor Rowing Championship</h2>
        </div>

        <div class="w3-row">
        {{ $i := 0}}
        {{ $cnt := 0 }}
        {{ range .Events }}
            {{ if .Entries }}
            <div class="w3-container w3-mobile w3-col s12 m6 l6 w3-margin-bottom w3-margin-top">
                <div class="w3-container w3-blue w3-round" id="event{{.ID}}">
                    <div class="w3-row">
                    <div class="w3-left w3-cell">Event {{ .ID }}</div>
                    <div class="w3-right w3-cell">{{ .Name }}</div>
                    </div><div class="w3-row">
                    <div class="w3-cell w3-left">{{ .Start }}</div>
                    <div class="w3-right w3-cell">{{ official . }} </div>                   
                    </div>
                </div>
                <div class="w3-row">
                        <div class="w3-col  w3-center s2"><b>Place</b></div>
                        <div class="w3-col w3-center s2"><b>Team</b></div>
                        <div class="w3-col  s6"><b>Name</b></div>
                        <div class="w3-col  s2 w3-center"><b>Time</b></div>
                </div>
                {{ range $place,$entry := .Entries }}
                    <div class="w3-row" id="{{$entry.ID}}">
                        <div class="w3-col  w3-center s2">{{ place $entry }}</div>
                        <div class="w3-col w3-center s2"> {{ $entry.ClubAbbrev }} </div>
                        <div class="w3-col  s6"> {{ $entry.BoatName }} {{ltwt $entry}}</div>
                        <div class="w3-col  s2 w3-center">{{ time $entry }}</div>    
                    </div>
                {{ end }}
            </div>
            {{$i = inc $i}}
            {{$cnt = inc $cnt }}
                {{ if needBreak $i }}
                    <div class="w3-cell w3-col s12 m12 l12"></div>
                {{ end }}
            {{ end }}
        {{ end }}
        </div>

    <div class="w3-col s12 w3-cell w3-center">Last updated on {{ now }}.</div>

    </body>
</html>
`
var scheduleTemplate = `
<html>
    <head>
        <title>
            2019 Cincinnati Indoor Rowing Championship Schedule
        </title>
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <link rel="stylesheet" href="https://www.w3schools.com/w3css/4/w3.css">
        <script src="https://code.jquery.com/jquery-3.3.1.min.js" integrity="sha256-FgpCb/KJQlLNfOu91ta32o/NMZxltwRo8QtmkMRdAu8="
        crossorigin="anonymous"></script>
    </head>

    <body>

        <div class="w3-container">
            <h2>2019 Cincinnati Indoor Rowing Championship</h2>
        </div>

        <div class="w3-cell-row">
        {{ $i := 0}}
        {{ $cnt := 0 }}
        {{ range .Races }}
            {{ if .Entries }}
                <div class="w3-cell w3-container w3-half w3-margin-bottom w3-margin-top">
                    <div class="w3-container w3-blue w3-round race" id="race{{.ID}}">
                        <div class="w3-row">
                        <div class="w3-left w3-cell">{{ .Start }}</div>
                        <div class="w3-right w3-cell">{{ .Name }}</div>
                        </div><div class="w3-row">
                        <div class="w3-cell w3-left">Race {{ .ID }}, Bank '{{ .Bank }}'</div>
                        <div class="w3-right w3-cell">{{ .Distance }} meters </div>                   
                        </div>
                    </div>
                    <div class="w3-cell-row">
                            <div class="w3-col  w3-center s2"><b>Lane</b></div>
                            <div class="w3-col w3-center s2"><b>Bow</b></div>
                            <div class="w3-col  s6"><b>Name</b></div>
                            <div class="w3-col  s2 w3-center"><b>Team</b></div>
                    </div>
                    {{ range .Entries }}
                        {{ $cnt = inc $cnt }}
                        <div class="w3-cell-row entry" id="{{.ID}}">
                            <div class="w3-col  w3-center s2">{{ .Lane }}</div>
                            <div class="w3-col w3-center s2">{{.ID}}</div>
                            <div class="w3-col   s6">{{ .BoatName }} {{ltwt .}}</div>
                            <div class="w3-col   s2 w3-center">{{ .ClubAbbrev }}</div>
                        </div>
                    {{ end }}
                </div>
                {{$i = inc $i}}
                {{ if needBreak $i }}
                    <div class="w3-cell w3-col s12 m12 l12"></div>
                {{ end }}
            {{ end }}
        {{ end }}
        </div>

        <div class="w3-col s12 w3-cell w3-center">{{ $cnt }} boats scheduled.  Last updated on {{ now }}.</div>

        <script>
            var hilightColor = "w3-green"
            
            function getHash() {
                return window.location.hash
            }

            function showHash() {
                $('.race').removeClass(hilightColor)
                $('.entry').removeClass(hilightColor)

                var id = getHash()

                if(id) {
                    $(id).addClass(hilightColor)
                }
            }

            $(function () {
                showHash()

                $(window).on('hashchange', function (e) {
                    showHash()
                });

            })
        </script>

    </body>
</html>
`

var noCreateTables bool

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

		// if user specifies not to create tables, we don't mess with the database at all
		if !noCreateTables {
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
		}
		// Create the new regatta
		fmt.Println("creating default config")
		C.WriteToFile("race.yaml")

		fmt.Println("creating shared folders")
		if err := createFolders(); err != nil {
			fmt.Printf("Can't create race folders: %s\n", err)
		}

		fmt.Println("creating html templates")
		if err := createTemplates(); err != nil {
			fmt.Println("Can't create html templates:", err)
		}
	},
}

func createTemplates() error {
	resultsFilename := filepath.Join(C.TemplatePath, "results.html")
	scheduleFilename := filepath.Join(C.TemplatePath, "schedule.html")

	if err := ioutil.WriteFile(resultsFilename, []byte(resultsTemplate), 0644); err != nil {
		return err
	}

	return ioutil.WriteFile(scheduleFilename, []byte(scheduleTemplate), 0644)
}

func createFolders() error {
	if err := os.MkdirAll(C.TemplatePath, 0755); err != nil {
		return err
	}

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
	var schema []string

	schema = append(schema, model.EntrySchema...)
	schema = append(schema, model.ResultSchema...)
	schema = append(schema, model.RaceSchema...)

	db := DBMustConnect()

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
	newCmd.Flags().BoolVar(&noCreateTables, "no-create-tables", false, "Don't create Database tables for a new race")
}
