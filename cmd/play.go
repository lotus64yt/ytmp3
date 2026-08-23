package cmd

import (
	"fmt"
	"net/url"
	"strings"
	"ytmp3/utils/youtube"

	"github.com/spf13/cobra"
)

var isSearch bool
var loop bool

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Play your video",
	Long:  `Play your video, use --background to disable console output`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")

		var videoID string

		if isSearch {
			arr, err := youtube.SearchVideos(query)
			if err != nil {
				fmt.Printf("An error occured during the search : %s", err)
				return
			}

			if len(arr) == 0 {
				fmt.Print("No results")
				return
			}

			videoID = arr[0].ID
		} else {
			if strings.HasPrefix(query, "http") {
				u, err := url.Parse(query)
				if err != nil {
					fmt.Printf("An error occured while parsing your url : %s", err)
					return
				}

				if !(strings.Contains(u.Path, "/watch") && u.Query().Has("v")) {
					fmt.Print("Invalid youtube url.")
					return
				}

				videoID = u.Query().Get("v")
			} else {
				videoID = query
			}
		}

		fmt.Printf("Playing %s...\n", videoID)
		video := youtube.Video{
			ID: videoID,
		}

		controls := youtube.PlayerControls{
			Loop: loop,
		}

		c, err := youtube.PlayAudio(video, controls)
		if err != nil {
			fmt.Println("Error :", err)
			return
		}

		fmt.Scanln()

		if err := c.Process.Kill(); err != nil {
			fmt.Println("Error while stopping the audio:", err)
		} else {
			fmt.Println("Stoped.")
		}
	},
}

func init() {
	rootCmd.AddCommand(playCmd)

	playCmd.Flags().BoolVarP(&isSearch, "search", "s", false, "Play the first video of this query. Usage: `ytmp3 play 'Never gonna give you up' --search`")
	playCmd.Flags().BoolVarP(&loop, "loop", "l", false, "Loop the playing video")
}
