package cmd

import (
	"discord-delete/client"
	"discord-delete/client/token"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var verbose bool
var dryrun bool
var skipChannels []string

var rootCmd = &cobra.Command{
	Use:   "discord-delete",
	Short: "A tool to delete Discord message history",
}

var partialCmd = &cobra.Command{
	Use: "partial",
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			log.SetLevel(log.DebugLevel)
		}

		log.Warn("Any tool that deletes your messages, including this one, could result in the termination of your account")
		log.Warn("You have been warned!")

		var tok string
		var err error

		tok, def := os.LookupEnv("DISCORD_TOKEN")

		if !def {
			tok, err = token.GetToken()
			if err != nil {
				log.Debug(err)
				log.Fatal("Error retrieving token, pass DISCORD_TOKEN as an environment variable instead")
			}
		}

		client := client.New(tok)

		client.SetDryRun(dryrun)
		client.SetSkipChannels(skipChannels)

		err = client.PartialDelete()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	partialCmd.Flags().BoolVarP(&dryrun, "dry-run", "d", false, "perform dry run without deleting anything")
	partialCmd.Flags().StringSliceVarP(&skipChannels, "skip", "s", []string{}, "skip message deletion for specified channels/guilds")

	rootCmd.AddCommand(partialCmd)
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
