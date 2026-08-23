package cmd

import (
	"fmt"
	"strings"
	"ytmp3/utils"
	"ytmp3/utils/youtube"

	"github.com/spf13/cobra"
)

var maxRes int

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for videos on YouTube",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")

		videos, err := youtube.SearchVideos(query)
		if err != nil {
			fmt.Printf("An error occurred: %s\n", err)
			return
		}

		videos = utils.ArrayCrop(videos, 0, maxRes)

		if len(videos) == 0 {
			fmt.Println("No videos found.")
			return
		}

		for i, v := range videos {
			fmt.Printf("[%d] %s (%s)\n", i+1, v.Title, v.Duration)
			fmt.Printf("    ID: %s | Chaîne: %s\n", v.ID, v.Author)
			fmt.Printf("    Lien: https://www.youtube.com/watch?v=%s\n\n", v.ID)
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().IntVarP(&maxRes, "max-result", "m", 10, "The maximum results output.")
}
