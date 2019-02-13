// Copyright © 2019 CJRC, Inc <greg@jrc.us>
//

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var dbConn string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "race",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./race.yaml)")
	rootCmd.PersistentFlags().StringVar(&dbConn, "db", "", "Connection string for the postgres database")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in pwd directory with name "race.yml"
		viper.AddConfigPath(pwd)
		viper.SetConfigName("race")
	}

	viper.AutomaticEnv() // read in environment variables that match
	viper.BindEnv("DB")  // need the db before init has been run and config file generated
	viper.SetDefault("DB", "")
	viper.SetDefault("HTMLPath", "shared/html")
	viper.SetDefault("RacePath", "shared/races")
	viper.SetDefault("ResultsPath", "shared/results")
	viper.SetEnvPrefix("RACE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(&C); err != nil {
		fmt.Printf("Fatal error config file: %s\n", err)
		os.Exit(1)
	}

	// The command line flag overrides the env variable or config file
	if dbConn != "" {
		C.DB = dbConn
	}
}
