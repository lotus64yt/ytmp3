package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
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

		fmt.Printf("Fetching info for %s...\n", videoID)
		video, err := youtube.GetVideoInfo(videoID)
		if err != nil {
			fmt.Println("Error getting video info:", err)
			return
		}

		controls := youtube.PlayerControls{
			Loop: loop,
		}

		c, err := youtube.PlayAudio(video, controls)
		if err != nil {
			fmt.Println("Error :", err)
			return
		}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		done := make(chan error, 1)
		go func() {
			done <- c.Wait()
		}()

		startUI(video, c, done, sigChan)

		fmt.Println("\nStopped.")
	},
}

func formatTime(seconds int) string {
	m := seconds / 60
	s := seconds % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

type MpvState struct {
	TimePos float64
	Paused  bool
}

func getMpvState() (MpvState, error) {
	var state MpvState
	state.TimePos = -1

	conn, err := net.DialTimeout("unix", "/tmp/ytmp3_ipc.sock", 50*time.Millisecond)
	if err != nil {
		return state, err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(50 * time.Millisecond))

	req := `{"command": ["get_property", "time-pos"], "request_id": 1}` + "\n" +
		`{"command": ["get_property", "pause"], "request_id": 2}` + "\n"
	conn.Write([]byte(req))

	scanner := bufio.NewScanner(conn)
	gotTime := false
	gotPause := false

	for scanner.Scan() {
		var resp struct {
			RequestID int         `json:"request_id"`
			Data      interface{} `json:"data"`
			Error     string      `json:"error"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &resp); err == nil {
			if resp.Error == "success" {
				if resp.RequestID == 1 {
					if v, ok := resp.Data.(float64); ok {
						state.TimePos = v
						gotTime = true
					}
				} else if resp.RequestID == 2 {
					if v, ok := resp.Data.(bool); ok {
						state.Paused = v
						gotPause = true
					}
				}
			}
		}
		if gotTime && gotPause {
			break
		}
	}

	if !gotTime && !gotPause {
		return state, fmt.Errorf("timeout or error")
	}
	return state, nil
}

func startUI(video youtube.Video, cmd *exec.Cmd, done chan error, sigChan chan os.Signal) {
	var currentSeconds int
	isPaused := false

	emptyBar := strings.Repeat("-", 30)
	fmt.Printf("\r\033[K\033[1;36m▶ %s\033[0m [%s] %s / %s",
		video.Title, emptyBar, formatTime(0), formatTime(video.Duration))

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			_ = cmd.Process.Kill()
			return
		case <-done:
			return
		case <-ticker.C:
			state, err := getMpvState()
			if err == nil {
				if state.TimePos >= 0 {
					currentSeconds = int(state.TimePos)
				}
				isPaused = state.Paused
			}

			progress := 0.0
			if video.Duration > 0 {
				progress = float64(currentSeconds) / float64(video.Duration)
			}
			if progress > 1.0 {
				progress = 1.0
			}

			filledUnits := int(progress * 30.0)
			emptyUnits := 30 - filledUnits

			var slider string
			if filledUnits > 0 {
				slider = strings.Repeat("=", filledUnits-1) + ">" + strings.Repeat("-", emptyUnits)
			} else {
				slider = strings.Repeat("-", 30)
			}

			icon := "▶"
			if isPaused {
				icon = "⏸"
			}

			fmt.Printf("\r\033[K\033[1;36m%s %s\033[0m [\033[1;38;2;215;170;140m%s\033[0m] %s / %s",
				icon, video.Title, slider, formatTime(currentSeconds), formatTime(video.Duration))
		}
	}
}

func init() {
	rootCmd.AddCommand(playCmd)

	playCmd.Flags().BoolVarP(&isSearch, "search", "s", false, "Play the first video of this query. Usage: `ytmp3 play 'Never gonna give you up' --search`")
	playCmd.Flags().BoolVarP(&loop, "loop", "l", false, "Loop the playing video")
}
