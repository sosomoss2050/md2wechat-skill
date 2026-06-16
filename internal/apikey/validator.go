package apikey

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrRequired     = errors.New("API_KEY_REQUIRED: MD2WECHAT_API_KEY is required")
	ErrInvalid      = errors.New("API_KEY_INVALID: MD2WECHAT_API_KEY is invalid")
	ErrVerifyFailed = errors.New("API_KEY_VERIFY_FAILED: failed to verify MD2WECHAT_API_KEY")
)

type Validator struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewValidator(baseURL, apiKey string) *Validator {
	return NewValidatorWithTimeout(baseURL, apiKey, 10*time.Second)
}

func NewValidatorWithTimeout(baseURL, apiKey string, timeout time.Duration) *Validator {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Validator{
		baseURL: strings.TrimSpace(baseURL),
		apiKey:  strings.TrimSpace(apiKey),
		client:  &http.Client{Timeout: timeout},
	}
}

func (v *Validator) Validate(ctx context.Context) error {
	if v == nil || v.apiKey == "" {
		return ErrRequired
	}
	endpoint, err := authValidateURL(v.baseURL)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrVerifyFailed, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, endpoint, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrVerifyFailed, err)
	}
	req.Header.Set("Authorization", "Bearer "+v.apiKey)

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrVerifyFailed, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return ErrInvalid
}

func authValidateURL(base string) (string, error) {
	if base == "" {
		base = "https://www.md2wechat.cn"
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	u.Path = strings.TrimRight(u.Path, "/")
	if strings.HasSuffix(u.Path, "/api") {
		u.Path += "/auth/validate"
	} else {
		u.Path += "/api/auth/validate"
	}
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

func IsRequired(err error) bool {
	return errors.Is(err, ErrRequired)
}

func IsInvalid(err error) bool {
	return errors.Is(err, ErrInvalid)
}

func IsVerifyFailed(err error) bool {
	return errors.Is(err, ErrVerifyFailed)
}
