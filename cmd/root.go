package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "ccs",
		Short: "CCS means central control service",
		Long:  `CCS means central control service.`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {

	rootCmd.AddCommand(NewFSMCmd())
	rootCmd.AddCommand(NewMessageServerCmd())
	// TODO add other cmds here
	rootCmd.AddCommand(NewClientCmd())
}
