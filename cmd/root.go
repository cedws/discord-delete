package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cedws/discord-delete/client"
	"github.com/cedws/discord-delete/client/token"
)

var (
	verbose      bool
	dryRun       bool
	skipPinned   bool
	minAge       uint
	maxAge       uint
	skipChannels []string
)

var rootCmd = &cobra.Command{
	Use:   "discord-delete",
	Short: "A tool to delete Discord message history",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose {
			log.SetLevel(log.DebugLevel)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Warn("any tool that deletes your messages, including this one, could result in the termination of your account")

		var tok string
		var err error

		tok, def := os.LookupEnv("DISCORD_TOKEN")
		if !def {
			if tok, err = token.GetToken(); err != nil {
				log.Debug(err)
				log.Fatal("error retrieving token, pass DISCORD_TOKEN as an environment variable instead")
			}
		}

		client := client.New(tok)
		client.SetDryRun(dryRun)
		client.SetSkipChannels(skipChannels)
		client.SetSkipPinned(skipPinned)

		if dryRun {
			log.Infof("no messages will be deleted in dry-run mode")
		}

		if minAge > 0 {
			if err := client.SetMinAge(minAge); err != nil {
				log.Fatal(err)
			}
			log.Infof("deleting messages older than %v days", minAge)
		}

		if maxAge > 0 {
			if err = client.SetMaxAge(maxAge); err != nil {
				log.Fatal(err)
			}
			log.Infof("deleting messages newer than %v days", maxAge)
		}

		if err = client.Delete(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "perform dry run without deleting anything")
	rootCmd.Flags().UintVarP(&minAge, "older-than-days", "o", 0, "minimum number in days of messages to be deleted")
	rootCmd.Flags().UintVarP(&maxAge, "newer-than-days", "n", 0, "maximum number in days of messages to be deleted")
	rootCmd.Flags().StringSliceVarP(&skipChannels, "skip", "s", []string{}, "skip message deletion for specified channels/guilds")
	rootCmd.Flags().BoolVarP(&skipPinned, "skip-pinned", "p", false, "skip message deletion for pinned messages")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
