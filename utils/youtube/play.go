package youtube

import (
	"fmt"
	"os/exec"
)

type Video struct {
	ID    string
	Title string
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
	}

	if controls.Loop {
		args = append(args, "--loop")
	}

	args = append(args, url)
	cmd := exec.Command("mpv",
		args...,
	)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("cant start mpv: %w", err)
	}

	return cmd, nil
}
