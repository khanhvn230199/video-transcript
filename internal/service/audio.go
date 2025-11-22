package service

import (
	"os/exec"
)

// ExtractAudio extracts audio from video to MP3 format
func ExtractAudio(videoPath, audioPath string) error {
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vn",                   // No video
		"-acodec", "libmp3lame", // MP3 codec
		"-b:a", "192k", // Audio bitrate 192kbps (chất lượng tốt)
		"-ar", "16000", // Sample rate 16kHz (Whisper khuyến nghị)
		"-ac", "1", // Mono
		audioPath,
		"-y",
	)
	return cmd.Run()
}
