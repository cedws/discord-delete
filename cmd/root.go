package cmd

import (
	"discord-delete/discord"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var Verbose bool

var rootCmd = &cobra.Command{
	Use:   "discord-delete",
	Short: "A tool to delete Discord message history",
}

var partialCmd = &cobra.Command{
	Use: "partial",
	Run: func(cmd *cobra.Command, args []string) {
		discord := discord.Client{
			Token: os.Getenv("DISCORD_TOKEN"),
		}
		discord.Me()
	},
}

func init() {
	rootCmd.AddCommand(partialCmd)
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "enable verbose logging")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
