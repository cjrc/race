// Copyright © 2019 CJRC, Inc <greg@jrc.us>
//

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// tallyCmd represents the tally command
var tallyCmd = &cobra.Command{
	Use:   "tally",
	Short: "Watch the results path and tally results as they arrive",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("tally called")
	},
}

func init() {
	rootCmd.AddCommand(tallyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tallyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tallyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
