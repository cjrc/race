// Copyright © 2019 CJRC, Inc <greg@jrc.us>
//

package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/cjrc/race/model"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// WriteGo is a flaf that determines if config should be dumped as GoLang or YAML
var WriteGo bool

// Config represents the global configuration
type Config struct {
	NLanes int // number of ergs per bank

	DB string `mapstructure:"DB"`

	HTMLPath     string
	RacePath     string
	ResultsPath  string
	TemplatePath string

	EntryCols struct { // Column numbers for the RegattaCentral generic boats.xls file
		EventID, BoatID, Age, Email, ClubName, ClubAbbrev, Seed, BoatName, Country int
	}

	MaxEntries int // Maximum number of lines that will be read from entries.xls

	Events []model.Event // The events in this regatta
}

// ConfigDefaults are passed to Viper to set the default config values
var ConfigDefaults = map[string]interface{}{
	"DB":                   "",
	"HTMLPath":             "shared/html",
	"RacePath":             "shared/races",
	"ResultsPath":          "shared/results",
	"TemplatePath":         "templates",
	"EntryCols.EventID":    0,
	"EntryCols.BoatID":     10,
	"EntryCols.Age":        12,
	"EntryCols.Email":      5,
	"EntryCols.ClubName":   8,
	"EntryCols.ClubAbbrev": 9,
	"EntryCols.Seed":       11,
	"EntryCols.BoatName":   14,
	"EntryCols.Country":    24,
	"MaxEntries":           2000,
	"Events": []model.Event{
		model.Event{ID: 1, Start: "8:00AM", Name: "Masters Men Age 30-39", Distance: 2000, Bank: "A", Entries: []model.Entry(nil)},
		model.Event{ID: 2, Start: "8:15AM", Name: "Masters Women Age 30-39", Distance: 2000, Bank: "B", Entries: []model.Entry(nil)},
		model.Event{ID: 3, Start: "8:00AM", Name: "Senior Men Age 40-49", Distance: 2000, Bank: "A", Entries: []model.Entry(nil)},
		model.Event{ID: 4, Start: "8:15AM", Name: "Senior Women Age 40-49", Distance: 2000, Bank: "B", Entries: []model.Entry(nil)},
		model.Event{ID: 5, Start: "8:00AM", Name: "Veteran Men Age 50+", Distance: 2000, Bank: "A", Entries: []model.Entry(nil)},
		model.Event{ID: 6, Start: "8:15AM", Name: "Veteran Women Age 50+", Distance: 2000, Bank: "B", Entries: []model.Entry(nil)},
		model.Event{ID: 7, Start: "8:00AM", Name: "Open Men", Distance: 2000, Bank: "A", Entries: []model.Entry(nil)},
		model.Event{ID: 8, Start: "8:15AM", Name: "Open Women", Distance: 2000, Bank: "B", Entries: []model.Entry(nil)},
		model.Event{ID: 9, Start: "8:30AM", Name: "Adaptive Men and Women", Distance: 1000, Bank: "A", Entries: []model.Entry(nil)},
		model.Event{ID: 10, Start: "8:45AM", Name: "Col. Novice Men", Distance: 2000, Bank: "A", Entries: []model.Entry(nil)},
		model.Event{ID: 11, Start: "8:53AM", Name: "Col. Novice Women", Distance: 2000, Bank: "B", Entries: []model.Entry(nil)},
		model.Event{ID: 12, Start: "9:45AM", Name: "Col. Varsity Men", Distance: 2000, Bank: "A", Entries: []model.Entry(nil)},
		model.Event{ID: 13, Start: "9:53AM", Name: "Col. Varsity Women", Distance: 2000, Bank: "B", Entries: []model.Entry(nil)},
		model.Event{ID: 14, Start: "10:45AM", Name: "Col. Coxswain Men", Distance: 1000, Bank: "A", Entries: []model.Entry(nil)},
		model.Event{ID: 15, Start: "10:53AM", Name: "Col. Coxswain Women", Distance: 1000, Bank: "B", Entries: []model.Entry(nil)},
		model.Event{ID: 16, Start: "11:30AM", Name: "JROW Boys and Girls", Distance: 1000, Bank: "A", Entries: []model.Entry(nil)},
		model.Event{ID: 17, Start: "11:45AM", Name: "HS Novice Boys", Distance: 2000, Bank: "A", Entries: []model.Entry(nil)},
		model.Event{ID: 18, Start: "11:53AM", Name: "HS Novice Girls", Distance: 2000, Bank: "B", Entries: []model.Entry(nil)},
		model.Event{ID: 19, Start: "12:30PM", Name: "HS Varsity Boys", Distance: 2000, Bank: "A", Entries: []model.Entry(nil)},
		model.Event{ID: 20, Start: "12:38PM", Name: "HS Varsity Girls", Distance: 2000, Bank: "B", Entries: []model.Entry(nil)},
		model.Event{ID: 21, Start: "1:30PM", Name: "HS Coxswain Boys", Distance: 1000, Bank: "A", Entries: []model.Entry(nil)},
		model.Event{ID: 22, Start: "1:38PM", Name: "HS Coxswain Girls", Distance: 1000, Bank: "B", Entries: []model.Entry(nil)},
	},
}

// C contains global configuration
var C Config

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Dumps the current configuration to stdout",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if WriteGo {
			C.WriteGo(os.Stdout)
		} else {
			C.WriteYAML(os.Stdout)
		}
	},
}

// WriteYAML writes the current config as YAML
func (config Config) WriteYAML(writer io.Writer) error {
	encoder := yaml.NewEncoder(writer)
	return encoder.Encode(&C)
}

// WriteGo writes the current config as GO lang
func (config Config) WriteGo(writer io.Writer) error {
	_, err := fmt.Fprintf(writer, "%#v\n", C)
	return err
}

// WriteToFile saves the configuration to the specified file
func (config Config) WriteToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return config.WriteYAML(file)
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	configCmd.Flags().BoolVar(&WriteGo, "go", false, "Dump the config as golang instead of YAML")
}
