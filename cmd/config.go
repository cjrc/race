// Copyright © 2019 CJRC, Inc <greg@jrc.us>
//

package cmd

import (
	"io"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/spf13/cobra"
)

// Config represents the global configuration
type Config struct {
	NLanes int // number of ergs per bank

	DB string `mapstructure:"DB"`

	HTMLPath    string
	RacePath    string
	ResultsPath string
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
