/*
Copyright © 2022 Pery Lemke <pery.lemke@gmail.com>

*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "timebench",
	Short: "A CLI to generate statistics to a TimescaleDB.",
	Long: `Timebench is a CLI library for Go that generate a benchmark statistics.
This application is a tool to generate a statistics on TimescaleDB.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
