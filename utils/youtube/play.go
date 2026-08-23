package youtube

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type Video struct {
	ID       string
	Title    string
	Duration int
	Channel  string
}

func GetVideoInfo(id string) (Video, error) {
	url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", id)
	cmd := exec.Command("yt-dlp", "--dump-json", url)
	out, err := cmd.Output()
	if err != nil {
		return Video{}, fmt.Errorf("failed to get video info: %w", err)
	}

	var info struct {
		Title    string `json:"title"`
		Duration int    `json:"duration"`
		Channel  string `json:"channel"`
	}
	if err := json.Unmarshal(out, &info); err != nil {
		return Video{}, fmt.Errorf("failed to parse video info: %w", err)
	}

	return Video{
		ID:       id,
		Title:    info.Title,
		Duration: info.Duration,
		Channel:  info.Channel,
	}, nil
}

type PlayerControls struct {
	Loop bool
}

func PlayAudio(video Video, controls PlayerControls) (*exec.Cmd, error) {
	url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", video.ID)
	args := []string{
		"--no-video",
		"--ytdl-format=bestaudio",
		"--really-quiet",
		fmt.Sprintf("--force-media-title=%s", video.Title),
		"--msg-level=all=no",
		"--input-ipc-server=/tmp/ytmp3_ipc.sock",
	}

	if controls.Loop {
		args = append(args, "--loop")
	}

	args = append(args, url)
	cmd := exec.Command("mpv", args...)

	os.Remove("/tmp/ytmp3_ipc.sock")

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("cant start mpv: %w", err)
	}

	return cmd, nil
}
