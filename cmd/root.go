package cmd

import (
	"discord-delete/discord"
	"discord-delete/log"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "discord-delete",
	Short: "A tool to delete Discord message history",
	Run: func(cmd *cobra.Command, args []string) {
		log.Init(verbose)
	},
}

var partialCmd = &cobra.Command{
	Use: "partial",
	Run: func(cmd *cobra.Command, args []string) {
		token := os.Getenv("DISCORD_TOKEN")
		if token == "" {
			log.Logger.Fatal("You must specify a Discord auth token by passing DISCORD_TOKEN as an environment variable.")
		}
		discord := discord.Client{
			Token: token,
		}
		err := discord.PartialDelete()
		if err != nil {
			log.Logger.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(partialCmd)
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
