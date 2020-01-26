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
		err = client.PartialDelete()
		if err != nil {
			log.Fatal(err)
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
