package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type SubtitleEntry struct {
	StartTime string      `json:"startTime"`
	EndTime   string      `json:"endTime"`
	Text      string      `json:"text"`
	Words     []WordEntry `json:"words,omitempty"` // Word-level timestamps
}

type WordEntry struct {
	Word      string  `json:"word"`
	Start     float64 `json:"start"`
	End       float64 `json:"end"`
	StartTime string  `json:"startTime"`
	EndTime   string  `json:"endTime"`
}

type WhisperResponse struct {
	Text string `json:"text"`
}

type WhisperWord struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type WhisperSegment struct {
	ID               int           `json:"id"`
	Seek             int           `json:"seek"`
	Start            float64       `json:"start"`
	End              float64       `json:"end"`
	Text             string        `json:"text"`
	Tokens           []int         `json:"tokens"`
	Temperature      float64       `json:"temperature"`
	AvgLogprob       float64       `json:"avg_logprob"`
	CompressionRatio float64       `json:"compression_ratio"`
	NoSpeechProb     float64       `json:"no_speech_prob"`
	Words            []WhisperWord `json:"words,omitempty"` // Word-level timestamps
}

type WhisperVerboseResponse struct {
	Task     string           `json:"task"`
	Language string           `json:"language"`
	Duration float64          `json:"duration"`
	Text     string           `json:"text"`
	Segments []WhisperSegment `json:"segments"`
}

type ElevenLabsWord struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type ElevenLabsChunk struct {
	Start float64          `json:"start"`
	End   float64          `json:"end"`
	Text  string           `json:"text"`
	Words []ElevenLabsWord `json:"words"`
}

type ElevenLabsSTTResponse struct {
	Text   string            `json:"text"`
	Chunks []ElevenLabsChunk `json:"chunks"`
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Tạo thư mục uploads nếu chưa có
	if err := os.MkdirAll("uploads", 0755); err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	// Serve static files
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// Trang chủ
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// API transcript - nhận video và trả về transcript
	r.POST("/api/transcript", handleTranscript)

	// API lấy danh sách voices từ ElevenLabs
	r.GET("/api/elevenlabs/voices", getElevenLabsVoices)

	// API text-to-speech với ElevenLabs
	r.POST("/api/elevenlabs/tts", textToSpeechElevenLabs)

	// Serve uploaded files
	r.StaticFS("/uploads", http.Dir("uploads"))

	fmt.Println("Server running on http://localhost:8080")
	r.Run(":8080")
}

// handleTranscript - API handler để xử lý upload video và tạo transcript
func handleTranscript(c *gin.Context) {
	log.Println("handleTranscript")
	// Lấy video file
	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Lưu video file
	filename := filepath.Base(file.Filename)
	videoPath := filepath.Join("uploads", filename)
	if err := c.SaveUploadedFile(file, videoPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	transcriptProvider := c.PostForm("transcriptProvider")
	if transcriptProvider == "" {
		transcriptProvider = "auto"
	}

	// Thử extract subtitles từ video trước
	subtitlePath := strings.TrimSuffix(videoPath, filepath.Ext(videoPath)) + ".srt"
	var subtitles []SubtitleEntry

	errExtract := extractSubtitles(videoPath, subtitlePath)
	if errExtract == nil {
		// Có subtitle trong video, parse nó
		parsedSubtitles, err := parseSRT(subtitlePath)
		if err != nil {
			log.Printf("Failed to parse SRT: %v", err)
		} else {
			// Kiểm tra xem subtitle có thực sự có nội dung không (không phải "No subtitles available")
			if len(parsedSubtitles) > 0 && !isEmptySubtitle(parsedSubtitles) {
				subtitles = parsedSubtitles
				log.Printf("Successfully extracted %d subtitle entries from video", len(subtitles))
			} else {
				log.Printf("Extracted subtitle file but it appears to be empty or invalid")
			}
		}
	}

	// Nếu không có subtitle trong video, dùng Whisper để tạo transcript
	if len(subtitles) == 0 {
		var err error

		switch transcriptProvider {
		case "elevenlabs":
			log.Println("Using ElevenLabs for transcription")
			subtitles, err = transcribeWithElevenLabs(videoPath)
			if err != nil {
				log.Printf("ElevenLabs transcription failed: %v. Falling back to Whisper.", err)
				subtitles = nil
			}
			if len(subtitles) == 0 {
				log.Println("Falling back to Whisper after ElevenLabs attempt")
				subtitles, err = transcribeWithWhisper(videoPath)
			}
		default:
			log.Println("Using Whisper for transcription")
			subtitles, err = transcribeWithWhisper(videoPath)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":    "Failed to transcribe video: " + err.Error(),
				"videoUrl": "/uploads/" + filename,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"videoUrl":    "/uploads/" + filename,
		"subtitles":   subtitles,
		"subtitleUrl": "/uploads/" + filepath.Base(subtitlePath),
	})
}

// Extract subtitles từ video sử dụng ffmpeg
func extractSubtitles(videoPath, outputPath string) error {
	streamIndexes, err := getSubtitleStreamIndexes(videoPath)
	if err != nil {
		return fmt.Errorf("ffprobe failed: %w", err)
	}

	if len(streamIndexes) == 0 {
		log.Printf("ffprobe found no subtitle streams in %s", videoPath)
		return fmt.Errorf("no subtitle streams found")
	}

	log.Printf("Found %d subtitle stream(s): %v", len(streamIndexes), streamIndexes)

	for _, idx := range streamIndexes {
		mapArg := fmt.Sprintf("0:s:%d", idx)
		cmd := exec.Command("ffmpeg",
			"-loglevel", "error",
			"-y",
			"-i", videoPath,
			"-map", mapArg,
			"-c:s", "srt",
			outputPath,
		)

		if err := cmd.Run(); err != nil {
			log.Printf("Failed to extract subtitle stream %s: %v", mapArg, err)
			continue
		}

		if hasValidSubtitle(outputPath) {
			log.Printf("Successfully extracted subtitle from stream %s", mapArg)
			return nil
		}

		log.Printf("Stream %s extracted but not valid SRT, trying next stream", mapArg)
	}

	log.Printf("No valid subtitles found in video: %s", videoPath)
	return fmt.Errorf("no subtitles found in video")
}

// Lấy danh sách subtitle stream indexes bằng ffprobe
func getSubtitleStreamIndexes(videoPath string) ([]int, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "s",
		"-show_entries", "stream=index",
		"-of", "csv=p=0",
		videoPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	indexes := []int{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		idx, err := strconv.Atoi(line)
		if err != nil {
			continue
		}
		indexes = append(indexes, idx)
	}

	return indexes, nil
}

// Kiểm tra xem file subtitle có nội dung hợp lệ không
func hasValidSubtitle(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	content := strings.TrimSpace(string(data))
	if content == "" || len(content) < 10 {
		return false
	}

	// Kiểm tra xem có phải là "No subtitles available" không
	if strings.Contains(strings.ToLower(content), "no subtitles available") {
		return false
	}

	// Kiểm tra xem có format SRT cơ bản không (có timestamp)
	return strings.Contains(content, "-->")
}

// Parse file SRT
func parseSRT(path string) ([]SubtitleEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	entries := []SubtitleEntry{}

	// Split by double newlines
	blocks := strings.Split(content, "\n\n")

	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		lines := strings.Split(block, "\n")
		if len(lines) < 3 {
			continue
		}

		// Line 0: index (skip)
		// Line 1: timestamp
		// Line 2+: text
		timestamp := strings.TrimSpace(lines[1])
		text := strings.Join(lines[2:], " ")

		// Parse timestamp: 00:00:00,000 --> 00:00:01,000
		parts := strings.Split(timestamp, " --> ")
		if len(parts) != 2 {
			continue
		}

		startTime := strings.TrimSpace(parts[0])
		endTime := strings.TrimSpace(parts[1])

		// Convert SRT time format to seconds
		startSeconds := srtTimeToSeconds(startTime)
		endSeconds := srtTimeToSeconds(endTime)

		entries = append(entries, SubtitleEntry{
			StartTime: fmt.Sprintf("%.3f", startSeconds),
			EndTime:   fmt.Sprintf("%.3f", endSeconds),
			Text:      text,
		})
	}

	return entries, nil
}

// Convert SRT time format (HH:MM:SS,mmm) to seconds
func srtTimeToSeconds(timeStr string) float64 {
	// Replace comma with dot for parsing
	timeStr = strings.Replace(timeStr, ",", ".", 1)

	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0
	}

	var hours, minutes int
	var seconds float64

	fmt.Sscanf(parts[0], "%d", &hours)
	fmt.Sscanf(parts[1], "%d", &minutes)
	fmt.Sscanf(parts[2], "%f", &seconds)

	return float64(hours*3600+minutes*60) + seconds
}

// Kiểm tra xem subtitle có phải là empty subtitle không
func isEmptySubtitle(subtitles []SubtitleEntry) bool {
	if len(subtitles) == 0 {
		return true
	}
	// Kiểm tra nếu chỉ có 1 entry với text "No subtitles available"
	if len(subtitles) == 1 {
		text := strings.ToLower(strings.TrimSpace(subtitles[0].Text))
		return text == "no subtitles available"
	}
	return false
}

// Extract audio từ video thành MP3 để gửi đến Whisper
func extractAudio(videoPath, audioPath string) error {
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

// Transcribe video sử dụng OpenAI Whisper API
func transcribeWithWhisper(videoPath string) ([]SubtitleEntry, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not found in .env file")
	}

	// Extract audio từ video thành MP3
	audioPath := strings.TrimSuffix(videoPath, filepath.Ext(videoPath)) + ".mp3"
	defer os.Remove(audioPath) // Clean up audio file after use

	if err := extractAudio(videoPath, audioPath); err != nil {
		return nil, fmt.Errorf("failed to extract audio: %v", err)
	}

	// Mở audio file
	audioFile, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %v", err)
	}
	defer audioFile.Close()

	// Tạo multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add file
	part, err := writer.CreateFormFile("file", filepath.Base(audioPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %v", err)
	}
	_, err = io.Copy(part, audioFile)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %v", err)
	}

	// Add model
	writer.WriteField("model", "whisper-1")
	// Add response_format để lấy segments với timestamps
	writer.WriteField("response_format", "verbose_json")
	// Add timestamp_granularities để lấy word-level timestamps
	writer.WriteField("timestamp_granularities[]", "word")

	writer.Close()

	// Tạo HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", &requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Gửi request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Đọc response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Whisper API error (status %d): %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("whisper API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var whisperResp WhisperVerboseResponse
	if err := json.Unmarshal(body, &whisperResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Convert Whisper segments thành SubtitleEntry với word-level timestamps
	subtitles := make([]SubtitleEntry, 0, len(whisperResp.Segments))
	for _, segment := range whisperResp.Segments {
		entry := SubtitleEntry{
			StartTime: fmt.Sprintf("%.3f", segment.Start),
			EndTime:   fmt.Sprintf("%.3f", segment.End),
			Text:      strings.TrimSpace(segment.Text),
		}

		// Convert word-level timestamps nếu có
		if len(segment.Words) > 0 {
			words := make([]WordEntry, 0, len(segment.Words))
			for _, word := range segment.Words {
				words = append(words, WordEntry{
					Word:      word.Word,
					Start:     word.Start,
					End:       word.End,
					StartTime: fmt.Sprintf("%.3f", word.Start),
					EndTime:   fmt.Sprintf("%.3f", word.End),
				})
			}
			entry.Words = words
		}

		subtitles = append(subtitles, entry)
	}

	return subtitles, nil
}

// Transcribe video sử dụng ElevenLabs Speech-to-Text API
func transcribeWithElevenLabs(videoPath string) ([]SubtitleEntry, error) {
	apiKey := os.Getenv("ELEVENLABS_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ELEVENLABS_API_KEY not found in .env file")
	}

	// Extract audio từ video thành MP3
	audioPath := strings.TrimSuffix(videoPath, filepath.Ext(videoPath)) + "_elevenlabs.mp3"
	defer os.Remove(audioPath)

	if err := extractAudio(videoPath, audioPath); err != nil {
		return nil, fmt.Errorf("failed to extract audio: %v", err)
	}

	audioFile, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %v", err)
	}
	defer audioFile.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("file", filepath.Base(audioPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %v", err)
	}
	if _, err := io.Copy(part, audioFile); err != nil {
		return nil, fmt.Errorf("failed to copy audio data: %v", err)
	}

	writer.WriteField("model_id", "scribe_v2")
	writer.WriteField("timestamps", "true")
	writer.Close()

	req, err := http.NewRequest("POST", "https://api.elevenlabs.io/v1/speech-to-text", &requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("xi-api-key", apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ElevenLabs API error (status %d): %s", resp.StatusCode, string(body))
	}

	var sttResp ElevenLabsSTTResponse
	if err := json.Unmarshal(body, &sttResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	subtitles := []SubtitleEntry{}

	if len(sttResp.Chunks) > 0 {
		for _, chunk := range sttResp.Chunks {
			entry := SubtitleEntry{
				StartTime: fmt.Sprintf("%.3f", chunk.Start),
				EndTime:   fmt.Sprintf("%.3f", chunk.End),
				Text:      strings.TrimSpace(chunk.Text),
			}
			if len(chunk.Words) > 0 {
				words := make([]WordEntry, 0, len(chunk.Words))
				for _, word := range chunk.Words {
					words = append(words, WordEntry{
						Word:      word.Word,
						Start:     word.Start,
						End:       word.End,
						StartTime: fmt.Sprintf("%.3f", word.Start),
						EndTime:   fmt.Sprintf("%.3f", word.End),
					})
				}
				entry.Words = words
			}
			subtitles = append(subtitles, entry)
		}
	} else if strings.TrimSpace(sttResp.Text) != "" {
		// Không có chunks, dùng toàn bộ text
		entry := SubtitleEntry{
			StartTime: "0.000",
			EndTime:   "0.000",
			Text:      strings.TrimSpace(sttResp.Text),
		}
		subtitles = append(subtitles, entry)
	}

	return subtitles, nil
}

// ElevenLabs API types
type ElevenLabsVoice struct {
	VoiceID     string `json:"voice_id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description,omitempty"`
	PreviewURL  string `json:"preview_url,omitempty"`
}

type ElevenLabsVoicesResponse struct {
	Voices []ElevenLabsVoice `json:"voices"`
}

type TTSRequest struct {
	Text    string `json:"text"`
	VoiceID string `json:"voice_id"`
	ModelID string `json:"model_id,omitempty"`
}

// getElevenLabsVoices - Lấy danh sách voices từ ElevenLabs
func getElevenLabsVoices(c *gin.Context) {
	apiKey := os.Getenv("ELEVENLABS_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ELEVENLABS_API_KEY not found in .env file"})
		return
	}

	req, err := http.NewRequest("GET", "https://api.elevenlabs.io/v1/voices", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request: " + err.Error()})
		return
	}

	req.Header.Set("xi-api-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response: " + err.Error()})
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "ElevenLabs API error",
			"status": resp.StatusCode,
			"body":   string(body),
		})
		return
	}

	var voicesResp ElevenLabsVoicesResponse
	if err := json.Unmarshal(body, &voicesResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"voices": voicesResp.Voices})
}

// textToSpeechElevenLabs - Text-to-speech với ElevenLabs
func textToSpeechElevenLabs(c *gin.Context) {
	apiKey := os.Getenv("ELEVENLABS_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ELEVENLABS_API_KEY not found in .env file"})
		return
	}

	var req TTSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default voice ID nếu không có
	if req.VoiceID == "" {
		req.VoiceID = "21m00Tcm4TlvDq8ikWAM" // Default voice
	}

	// Default model
	if req.ModelID == "" {
		req.ModelID = "eleven_multilingual_v2"
	}

	// Tạo request body
	requestBody := map[string]interface{}{
		"text":     req.Text,
		"model_id": req.ModelID,
		"voice_settings": map[string]float64{
			"stability":        0.5,
			"similarity_boost": 0.75,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request: " + err.Error()})
		return
	}

	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", req.VoiceID)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request: " + err.Error()})
		return
	}

	httpReq.Header.Set("Accept", "audio/mpeg")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("xi-api-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "ElevenLabs API error",
			"status": resp.StatusCode,
			"body":   string(body),
		})
		return
	}

	// Đọc audio data
	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read audio data: " + err.Error()})
		return
	}

	// Lưu file audio
	filename := fmt.Sprintf("tts_%d.mp3", time.Now().Unix())
	audioPath := filepath.Join("uploads", filename)
	if err := os.WriteFile(audioPath, audioData, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save audio file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"audioUrl": "/uploads/" + filename,
		"message":  "Audio generated successfully",
	})
}
