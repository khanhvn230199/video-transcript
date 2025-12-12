package helper

import (
	"context"
	"fmt"

	api "github.com/deepgram/deepgram-go-sdk/v3/pkg/api/listen/v1/rest"
	interfacesv1 "github.com/deepgram/deepgram-go-sdk/v3/pkg/api/listen/v1/rest/interfaces"
	interfaces "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/interfaces"
	client "github.com/deepgram/deepgram-go-sdk/v3/pkg/client/listen"
	"go.uber.org/zap"
)

func DeepgramSTTFromBytes(ctx context.Context, file_url string, contentType string, language string) (*interfacesv1.PreRecordedResponse, error) {
	if language == "" {
		language = "en-US"
	}

	// 1. Init client Listen API
	options := &interfaces.PreRecordedTranscriptionOptions{
		Model:          "nova-3",
		Keyterm:        []string{"deepgram"},
		Punctuate:      true,
		Diarize:        true,
		Language:       language,
		Utterances:     true,
		Redact:         []string{"pci", "ssn"},
		DetectLanguage: true,
	}

	c := client.NewRESTWithDefaults()
	dg := api.New(c)

	res, err := dg.FromURL(ctx, file_url, options)
	if err != nil {
		if e, ok := err.(*interfaces.StatusError); ok {
			zap.S().Errorw("DEEPGRAM ERROR", "error", e.DeepgramError.ErrCode, "message", e.DeepgramError.ErrMsg)
			return nil, fmt.Errorf("DEEPGRAM ERROR: %w", err)
		}
		zap.S().Errorw("FromURL failed", "error", err)
		return nil, fmt.Errorf("FromURL failed: %w", err)
	}

	return res, nil
}
