package cmd

import (
	"discord-delete/discord"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

var verbose bool
var find bool

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

		token := os.Getenv("DISCORD_TOKEN")

		if find {
			appdata := os.Getenv("APPDATA")
			path := filepath.Join(appdata, "Discord/Local Storage/leveldb")

			// TODO: Database read fails if it's locked by the Discord client.
			db, err := leveldb.OpenFile(path, &opt.Options{
				ReadOnly: true,
			})
			if err != nil {
				log.Fatal("Couldn't retrieve client token, try logging into the client or passing DISCORD_TOKEN as an environment variable")
			}
			defer db.Close()

			data, err := db.Get([]byte("_https://discordapp.com\x00\x01token"), nil)
			if err != nil {
				log.Fatal("Couldn't retrieve client token, try logging into the client or passing DISCORD_TOKEN as an environment variable")
			}

			// TODO: Try to improve this expression or use capture groups.
			reg := regexp.MustCompile("\"(.*?)\"")
			token, err = strconv.Unquote(string(reg.Find(data)))
			if err != nil {
				log.Fatal(err)
			}
		}

		client := discord.New(token)

		err := client.PartialDelete()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(partialCmd)
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	rootCmd.PersistentFlags().BoolVarP(&find, "find", "f", false, "find the auth token automatically")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
