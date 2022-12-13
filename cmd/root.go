package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	// cfgFile     string
	// userLicense string

	rootCmd = &cobra.Command{
		Use:   "gphotos",
		Short: "CLI to interact with google photos",
		Long:  `gphotos is a CLI to interact with google photos.`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
