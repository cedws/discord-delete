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
	minAge       string
	maxAge       string
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

	if minAge != "" {
		err = client.SetMinAge(minAge)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Deleting messages with a minimum age of %v", minAge)
	}

	if maxAge != "" {
		err = client.SetMaxAge(maxAge)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Deleting messages with a maximum age of %v", maxAge)
	}

	err = client.PartialDelete()
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	partialCmd.Flags().BoolVarP(&dryrun, "dry-run", "d", false, "perform dry run without deleting anything")
	partialCmd.Flags().StringVarP(&minAge, "min-age", "i", "", "minimum age of messages to delete")
	partialCmd.Flags().StringVarP(&maxAge, "max-age", "a", "", "maximum age of messages to delete")
	partialCmd.Flags().StringSliceVarP(&skipChannels, "skip", "s", []string{}, "skip message deletion for specified channels/guilds")
}
