package cmd

import (
	"discord-delete/client"
	"discord-delete/client/token"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	dryrun       bool
	minAge       uint
	maxAge       uint
	skipChannels []string
)

var partialCmd = &cobra.Command{
	Use: "partial",
	Run: partial,
}

func partial(cmd *cobra.Command, args []string) {
	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	log.Warn("Any tool that deletes your messages, including this one, could result in the termination of your account. You have been warned!")

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
	
	if dryrun {
		log.Infof("No messages will be deleted in dry-run mode")
	}

	if minAge > 0 {
		err = client.SetMinAge(minAge)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Deleting messages older than %v days", minAge)
	}

	if maxAge > 0 {
		err = client.SetMaxAge(maxAge)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Deleting messages newer than %v days", maxAge)
	}

	err = client.PartialDelete()
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	partialCmd.Flags().BoolVarP(&dryrun, "dry-run", "d", false, "perform dry run without deleting anything")
	partialCmd.Flags().UintVarP(&minAge, "older-than-days", "o", 0, "minimum number in days of messages to be deleted")
	partialCmd.Flags().UintVarP(&maxAge, "newer-than-days", "n", 0, "maximum number in days of messages to be deleted")
	partialCmd.Flags().StringSliceVarP(&skipChannels, "skip", "s", []string{}, "skip message deletion for specified channels/guilds")
}
