package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootFlags = struct {
}{}

var rootCmd = &cobra.Command{
	Use:   "tci",
	Short: "A client for the TCI protocol.",
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
}
