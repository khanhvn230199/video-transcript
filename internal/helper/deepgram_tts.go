package helper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"video-transcript/internal/config"
	"video-transcript/internal/uploads"

	api "github.com/deepgram/deepgram-go-sdk/v3/pkg/api/speak/v1/rest"
	interfaces "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/interfaces"
	client "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/speak"
	"go.uber.org/zap"
)

func DeepgramTTS(ctx context.Context, userID string, text string) (string, error) {
	// Deepgram client is already initialized in main()
	if config.SvcCfg.DeepgramAPIKey == "" {
		err := errors.New("missing Deepgram API key")
		zap.S().Errorw("Deepgram client init failed", "error", err)
		return "", err
	}
	options := &interfaces.SpeakOptions{
		Model: "aura-2-thalia-en",
	}
	log.Println("Deepgram API key", config.SvcCfg.DeepgramAPIKey)
	c := client.NewRESTWithDefaults()
	if c == nil {
		log.Println("Deepgram client is not initialized", config.SvcCfg.DeepgramAPIKey)
		zap.S().Errorw("Deepgram client is not initialized", "error", errors.New("deepgram client is not initialized"))
		return "", fmt.Errorf("deepgram client is not initialized: %w", errors.New("deepgram client is not initialized"))
	}
	dg := api.New(c)
	if dg == nil {
		zap.S().Errorw("Deepgram API is not initialized", "error", errors.New("deepgram API is not initialized"))
		return "", fmt.Errorf("deepgram API is not initialized: %w", errors.New("deepgram API is not initialized"))
	}
	var buf bytes.Buffer
	res, err := dg.ToFile(ctx, text, options, &buf)
	if err != nil {
		zap.S().Errorw("from stream failed", "error", err)
		return "", fmt.Errorf("from stream failed: %w", err)
	}

	contentType := "application/octet-stream"
	if res != nil && res.ContextType != "" {
		contentType = res.ContextType
	}

	// Lưu thẳng audio bytes lên R2, không cần ghi ra file tạm.
	key := fmt.Sprintf("text-to-speech/%s/%d-audio", userID, time.Now().UnixNano())
	url, err := uploads.UploadToR2(ctx, key, bytes.NewReader(buf.Bytes()), int64(buf.Len()), contentType)
	if err != nil {
		zap.S().Error("UploadToR2 Err: %v", err)
		return "", fmt.Errorf("UploadToR2 Err: %v", err)
	}
	return url, nil
}
