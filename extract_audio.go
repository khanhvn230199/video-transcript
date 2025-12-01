package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
)

// extractAudio - Extract audio từ video thành MP3 để gửi đến API
func extractAudio(videoPath, audioPath string) error {
	log.Printf("🔧 Running ffmpeg to extract audio (MP3, 16kHz mono, 192kbps)...")
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vn",                   // No video
		"-acodec", "libmp3lame", // MP3 codec
		"-b:a", "192k", // Audio bitrate 192kbps (chất lượng tốt)
		"-ar", "16000", // Sample rate 16kHz (API khuyến nghị)
		"-ac", "1", // Mono
		"-loglevel", "error", // Chỉ hiển thị lỗi
		audioPath,
		"-y", // Overwrite output file
	)

	// Capture error output for better debugging
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ffmpeg execution failed: %v\nffmpeg stderr: %s", err, stderr.String())
	}
	return nil
}
