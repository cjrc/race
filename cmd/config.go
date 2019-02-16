// Copyright © 2019 CJRC, Inc <greg@jrc.us>
//

package cmd

import (
	"io"
	"os"

	"github.com/cjrc/race/model"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

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
		model.Event{ID: 1, Start: "8:00AM", Name: "Open Men", Distance: 2000, Bank: "A"},
		model.Event{ID: 2, Start: "8:15AM", Name: "Open Women", Distance: 2000, Bank: "B"},
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
		C.Write(os.Stdout)
	},
}

func (config Config) Write(writer io.Writer) error {
	encoder := yaml.NewEncoder(writer)
	return encoder.Encode(&C)
}

// WriteToFile saves the configuration to the specified file
func (config Config) WriteToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return config.Write(file)
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
}
