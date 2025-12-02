package helper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"video-transcript/internal/config"

	"go.uber.org/zap"
)

const (
	deepgramURL    = "https://api.deepgram.com/v1/listen?model=nova-2&smart_format=true"
	defaultRetries = 3
)

// deepgramRequest là payload gửi lên Deepgram.
type deepgramRequest struct {
	URL string `json:"url"`
}

// DeepgramWord là 1 từ trong transcript (có timestamp).
type DeepgramWord struct {
	Word           string  `json:"word"`
	Start          float64 `json:"start"`
	End            float64 `json:"end"`
	Confidence     float64 `json:"confidence"`
	PunctuatedWord string  `json:"punctuated_word"`
}

// DeepgramAlternative là 1 phương án transcript.
type DeepgramAlternative struct {
	Transcript string         `json:"transcript"`
	Confidence float64        `json:"confidence"`
	Words      []DeepgramWord `json:"words"`
}

// deepgramChannel tương ứng với 1 audio channel.
type deepgramChannel struct {
	Alternatives []DeepgramAlternative `json:"alternatives"`
}

// deepgramResults chứa danh sách channel.
type deepgramResults struct {
	Channels []deepgramChannel `json:"channels"`
}

// DeepgramResponse ánh xạ phần `results` quan trọng từ response Deepgram.
type DeepgramResponse struct {
	Results deepgramResults `json:"results"`
}

// TranscribeWithDeepgram gọi Deepgram để lấy transcript cho 1 URL audio, có retry.
func TranscribeWithDeepgram(ctx context.Context, audioURL string) (*DeepgramResponse, error) {
	if config.SvcCfg.DeepgramAPIKey == "" {
		zap.S().Error("Deepgram API key is not configured")
		return nil, fmt.Errorf("Deepgram API key is not configured")
	}

	bodyBytes, err := json.Marshal(deepgramRequest{URL: audioURL})
	if err != nil {
		zap.S().Error("marshal deepgram request: %w", err)
		return nil, fmt.Errorf("marshal deepgram request: %w", err)
	}

	var lastErr error
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	for attempt := 0; attempt < defaultRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, deepgramURL, bytes.NewReader(bodyBytes))
		if err != nil {
			zap.S().Error("create request: %w", err)
			return nil, fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Authorization", "Token "+config.SvcCfg.DeepgramAPIKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			zap.S().Error("call deepgram: %w", err)
			lastErr = fmt.Errorf("call deepgram: %w", err)
		} else {
			defer resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				respBytes, err := io.ReadAll(resp.Body)
				if err != nil {
					zap.S().Error("read response: %w", err)
					return nil, fmt.Errorf("read response: %w", err)
				}

				var dgResp DeepgramResponse
				if err := json.Unmarshal(respBytes, &dgResp); err != nil {
					zap.S().Error("unmarshal response: %w", err)
					return nil, fmt.Errorf("unmarshal response: %w", err)
				}

				return &dgResp, nil
			}

			// Với status code 5xx thì retry, còn lại trả lỗi luôn.
			if resp.StatusCode < 500 {
				respBytes, _ := io.ReadAll(resp.Body)
				zap.S().Error("deepgram non-retryable status %d: %s", resp.StatusCode, string(respBytes))
				return nil, fmt.Errorf("deepgram non-retryable status %d: %s", resp.StatusCode, string(respBytes))
			}

			zap.S().Error("deepgram status %d", resp.StatusCode)
			lastErr = fmt.Errorf("deepgram status %d", resp.StatusCode)
		}

		// exponential backoff đơn giản
		sleep := time.Duration(attempt+1) * time.Second
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(sleep):
		}
	}

	if lastErr == nil {
		zap.S().Error("deepgram request failed after %d retries", defaultRetries)
		lastErr = fmt.Errorf("deepgram request failed after %d retries", defaultRetries)
	}
	return nil, lastErr
}
