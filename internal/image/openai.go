package image

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/geekjourneyx/md2wechat-skill/internal/config"
)

// OpenAIProvider OpenAI 图片生成服务提供者
type OpenAIProvider struct {
	apiKey  string
	baseURL string
	model   string
	size    string
	client  *http.Client
}

// NewOpenAIProvider 创建 OpenAI Provider
func NewOpenAIProvider(cfg *config.Config) (*OpenAIProvider, error) {
	model := cfg.ImageModel
	if model == "" {
		model = DefaultProviderModel("openai")
	}

	size := cfg.ImageSize
	if size == "" {
		size = "auto"
	}

	return &OpenAIProvider{
		apiKey:  cfg.ImageAPIKey,
		baseURL: cfg.ImageAPIBase,
		model:   model,
		size:    size,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// Name 返回提供者名称
func (p *OpenAIProvider) Name() string {
	return "OpenAI"
}

// Generate 生成图片
func (p *OpenAIProvider) Generate(ctx context.Context, prompt string) (*GenerateResult, error) {
	// 构造请求
	reqBody := map[string]any{
		"model":  p.model,
		"prompt": prompt,
		"n":      1,
		"size":   p.size,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, &GenerateError{
			Provider: p.Name(),
			Code:     "marshal_error",
			Message:  "请求构造失败",
			Original: err,
		}
	}

	// 创建请求
	url := p.baseURL + "/images/generations"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, &GenerateError{
			Provider: p.Name(),
			Code:     "request_error",
			Message:  "创建请求失败",
			Original: err,
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	// 发送请求
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, &GenerateError{
			Provider: p.Name(),
			Code:     "network_error",
			Message:  "网络请求失败，请检查网络连接",
			Hint:     "确认网络连接正常，API 地址正确",
			Original: err,
		}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// 处理错误响应
	if resp.StatusCode != http.StatusOK {
		return nil, p.handleErrorResponse(resp)
	}

	// 解析响应
	var result struct {
		Data []struct {
			B64JSON       string `json:"b64_json,omitempty"`
			URL           string `json:"url"`
			RevisedPrompt string `json:"revised_prompt,omitempty"`
			OutputFormat  string `json:"output_format,omitempty"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, &GenerateError{
			Provider: p.Name(),
			Code:     "decode_error",
			Message:  "响应解析失败",
			Original: err,
		}
	}

	if len(result.Data) == 0 {
		return nil, &GenerateError{
			Provider: p.Name(),
			Code:     "no_image",
			Message:  "未生成图片",
			Hint:     "提示词可能不符合内容政策，请尝试修改提示词",
		}
	}

	image := result.Data[0]
	imagePath := strings.TrimSpace(image.URL)
	if imagePath == "" && strings.TrimSpace(image.B64JSON) != "" {
		imagePath, err = p.saveBase64Image(image.B64JSON, image.OutputFormat)
		if err != nil {
			return nil, err
		}
	}
	if imagePath == "" {
		return nil, &GenerateError{
			Provider: p.Name(),
			Code:     "no_image",
			Message:  "未生成图片",
			Hint:     "OpenAI 响应中没有可下载 URL 或 base64 图片数据，请检查模型与响应格式",
		}
	}

	return &GenerateResult{
		URL:           imagePath,
		RevisedPrompt: image.RevisedPrompt,
		Model:         p.model,
		Size:          p.size,
	}, nil
}

func (p *OpenAIProvider) saveBase64Image(b64, outputFormat string) (string, error) {
	imageData, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", &GenerateError{
			Provider: p.Name(),
			Code:     "decode_error",
			Message:  "OpenAI 图片数据解析失败",
			Original: err,
		}
	}
	if len(imageData) == 0 {
		return "", &GenerateError{
			Provider: p.Name(),
			Code:     "empty_data",
			Message:  "OpenAI 图片数据为空",
		}
	}

	var ext string
	switch strings.ToLower(strings.TrimSpace(outputFormat)) {
	case "jpeg", "jpg":
		ext = ".jpg"
	case "webp":
		ext = ".webp"
	case "png", "":
		ext = ".png"
	default:
		ext = ".png"
	}

	tmpFile, err := os.CreateTemp("", "md2wechat-openai-*"+ext)
	if err != nil {
		return "", &GenerateError{
			Provider: p.Name(),
			Code:     "write_error",
			Message:  "图片保存失败",
			Original: err,
		}
	}
	tmpPath := tmpFile.Name()
	if _, err := tmpFile.Write(imageData); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return "", &GenerateError{
			Provider: p.Name(),
			Code:     "write_error",
			Message:  "图片保存失败",
			Original: err,
		}
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", &GenerateError{
			Provider: p.Name(),
			Code:     "write_error",
			Message:  "图片保存失败",
			Original: err,
		}
	}

	return tmpPath, nil
}

// handleErrorResponse 处理错误响应
func (p *OpenAIProvider) handleErrorResponse(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var errResp struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		} `json:"error"`
	}

	// 尝试解析 OpenAI 错误格式
	_ = json.Unmarshal(body, &errResp)

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return &GenerateError{
			Provider: p.Name(),
			Code:     "unauthorized",
			Message:  "API Key 无效或已过期",
			Hint:     "请检查配置文件中的 api.image_key 是否正确",
			Original: fmt.Errorf("status 401: %s", string(body)),
		}
	case http.StatusTooManyRequests:
		return &GenerateError{
			Provider: p.Name(),
			Code:     "rate_limit",
			Message:  "请求过于频繁，请稍后重试",
			Hint:     "OpenAI API 有速率限制，请等待一段时间后再试",
			Original: fmt.Errorf("status 429: %s", string(body)),
		}
	case http.StatusBadRequest:
		hint := "请检查图片尺寸、模型名称等参数是否正确"
		if modelsHint := ProviderSupportedModelsHint("openai"); modelsHint != "" {
			hint += "。" + modelsHint
		}
		return &GenerateError{
			Provider: p.Name(),
			Code:     "bad_request",
			Message:  fmt.Sprintf("请求参数错误: %s", errResp.Error.Message),
			Hint:     hint,
			Original: fmt.Errorf("status 400: %s", string(body)),
		}
	case http.StatusPaymentRequired, http.StatusForbidden:
		return &GenerateError{
			Provider: p.Name(),
			Code:     "payment_required",
			Message:  "账户余额不足或访问受限",
			Hint:     "请检查 OpenAI 账户余额和 API 使用权限",
			Original: fmt.Errorf("status %d: %s", resp.StatusCode, string(body)),
		}
	default:
		return &GenerateError{
			Provider: p.Name(),
			Code:     "unknown",
			Message:  fmt.Sprintf("API 返回错误 (HTTP %d)", resp.StatusCode),
			Hint:     "请稍后重试，或检查 OpenAI 服务状态",
			Original: fmt.Errorf("status %d: %s", resp.StatusCode, string(body)),
		}
	}
}
