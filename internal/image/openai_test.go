package image

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/geekjourneyx/md2wechat-skill/internal/config"
)

func TestOpenAIProviderGenerateDecodesB64JSON(t *testing.T) {
	imageBytes := []byte{0x89, 0x50, 0x4e, 0x47}
	provider, err := NewOpenAIProvider(&config.Config{
		ImageAPIKey:  "image-key",
		ImageAPIBase: "https://api.openai.test/v1",
		ImageModel:   "gpt-image-2",
	})
	if err != nil {
		t.Fatalf("NewOpenAIProvider() error = %v", err)
	}
	provider.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/images/generations" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req["model"] != "gpt-image-2" {
			t.Fatalf("model = %v", req["model"])
		}
		body := `{"data":[{"b64_json":"` + base64.StdEncoding.EncodeToString(imageBytes) + `","output_format":"png"}]}`
		return jsonResponse(http.StatusOK, body), nil
	})}

	result, err := provider.Generate(context.Background(), "draw a fox")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	defer func() {
		_ = os.Remove(result.URL)
	}()

	if result.URL == "" || !strings.HasSuffix(result.URL, ".png") {
		t.Fatalf("URL = %q, want temp png path", result.URL)
	}
	got, err := os.ReadFile(result.URL)
	if err != nil {
		t.Fatalf("read generated temp file: %v", err)
	}
	if string(got) != string(imageBytes) {
		t.Fatalf("image bytes = %v, want %v", got, imageBytes)
	}
}

func TestOpenAIProviderGenerateRejectsEmptyImagePayload(t *testing.T) {
	provider, err := NewOpenAIProvider(&config.Config{
		ImageAPIKey:  "image-key",
		ImageAPIBase: "https://api.openai.test/v1",
		ImageModel:   "gpt-image-2",
	})
	if err != nil {
		t.Fatalf("NewOpenAIProvider() error = %v", err)
	}
	provider.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, `{"data":[{}]}`), nil
	})}

	_, err = provider.Generate(context.Background(), "draw a fox")
	if err == nil {
		t.Fatal("expected error for empty image payload")
	}
	genErr, ok := err.(*GenerateError)
	if !ok {
		t.Fatalf("error type = %T, want *GenerateError", err)
	}
	if genErr.Code != "no_image" {
		t.Fatalf("error code = %q, want no_image", genErr.Code)
	}
}
