// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var dbConn string

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
	},
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
	newCmd.PersistentFlags().String("db", "", "Connection string for the postgres database")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}