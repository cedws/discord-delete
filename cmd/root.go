package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	verbose bool
	rootCmd = &cobra.Command{
		Use:   "discord-delete",
		Short: "A tool to delete Discord message history",
	}
)

func init() {
	rootCmd.AddCommand(partialCmd)
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
