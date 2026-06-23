package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/geekjourneyx/md2wechat-skill/internal/config"
)

func TestPrepareWeChatSideEffectRequiresAPIKeyForNamedAccount(t *testing.T) {
	oldCfg, oldAccount := cfg, wechatAccountName
	oldValidate := validateAPIKeyForWeChatAccount
	t.Cleanup(func() {
		cfg = oldCfg
		wechatAccountName = oldAccount
		validateAPIKeyForWeChatAccount = oldValidate
	})

	cfg = &config.Config{
		WechatAccounts: map[string]config.WechatAccount{
			"main": {AppID: "appid", Secret: "secret"},
		},
	}
	wechatAccountName = "main"
	validateAPIKeyForWeChatAccount = func(apiKey string) error {
		t.Fatal("validator should not be called when key is missing")
		return nil
	}

	err := prepareWeChatSideEffect()
	if err == nil || !strings.Contains(err.Error(), "API_KEY_REQUIRED") {
		t.Fatalf("prepareWeChatSideEffect() error = %v", err)
	}
}

func TestPrepareWeChatSideEffectHonorsFlagAfterAmbiguousLoad(t *testing.T) {
	oldCfg, oldLog, oldAccount := cfg, log, wechatAccountName
	oldValidate := validateAPIKeyForWeChatAccount
	t.Cleanup(func() {
		cfg = oldCfg
		log = oldLog
		wechatAccountName = oldAccount
		validateAPIKeyForWeChatAccount = oldValidate
	})

	home := t.TempDir()
	configDir := filepath.Join(home, ".config", "md2wechat")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(strings.TrimSpace(`
wechat:
  accounts:
    main:
      appid: wx-main
      secret: secret-main
    client-a:
      appid: wx-client
      secret: secret-client
`)+"\n"), 0600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	t.Setenv("HOME", home)
	t.Setenv("WECHAT_APPID", "")
	t.Setenv("WECHAT_SECRET", "")
	t.Setenv("WECHAT_ACCOUNT", "")
	t.Setenv("MD2WECHAT_API_KEY", "")

	cfg = nil
	log = nil
	wechatAccountName = "main"
	validateAPIKeyForWeChatAccount = func(apiKey string) error {
		t.Fatal("validator should not be called when key is missing")
		return nil
	}

	if err := initConfig(); err != nil {
		t.Fatalf("initConfig() error = %v", err)
	}
	err := prepareWeChatSideEffect()
	cliErr, ok := err.(*cliError)
	if !ok || cliErr.Code != codeAPIKeyRequired {
		t.Fatalf("prepareWeChatSideEffect() error = %#v", err)
	}
	if cfg.WechatAppID != "wx-main" || cfg.WechatSecret != "secret-main" {
		t.Fatalf("selected credentials = %q/%q", cfg.WechatAppID, cfg.WechatSecret)
	}
}

func TestRunInspectFailsClosedForMissingExplicitWechatAccount(t *testing.T) {
	oldCfg, oldAccount := cfg, wechatAccountName
	oldUpload, oldDraft := inspectUpload, inspectDraft
	t.Cleanup(func() {
		cfg = oldCfg
		wechatAccountName = oldAccount
		inspectUpload = oldUpload
		inspectDraft = oldDraft
	})

	cfg = &config.Config{
		WechatAppID:  "direct-appid",
		WechatSecret: "direct-secret",
		WechatAccounts: map[string]config.WechatAccount{
			"main": {AppID: "wx-main", Secret: "secret-main"},
		},
		MD2WechatAPIKey: "api-key",
	}
	wechatAccountName = "missing"
	inspectDraft = true
	inspectUpload = false

	markdownPath := filepath.Join(t.TempDir(), "article.md")
	if err := os.WriteFile(markdownPath, []byte("# Title\n"), 0600); err != nil {
		t.Fatalf("write markdown: %v", err)
	}

	_, err := runInspect(markdownPath)
	cliErr, ok := err.(*cliError)
	if !ok || cliErr.Code != codeWechatAccountNotFound {
		t.Fatalf("runInspect() error = %#v", err)
	}
}

func TestPrepareWeChatSideEffectDoesNotValidateDirectLegacyPath(t *testing.T) {
	oldCfg, oldAccount := cfg, wechatAccountName
	oldValidate := validateAPIKeyForWeChatAccount
	t.Cleanup(func() {
		cfg = oldCfg
		wechatAccountName = oldAccount
		validateAPIKeyForWeChatAccount = oldValidate
	})

	cfg = &config.Config{WechatAppID: "appid", WechatSecret: "secret"}
	wechatAccountName = ""
	called := false
	validateAPIKeyForWeChatAccount = func(apiKey string) error {
		called = true
		return nil
	}

	if err := prepareWeChatSideEffect(); err != nil {
		t.Fatalf("prepareWeChatSideEffect() error = %v", err)
	}
	if called {
		t.Fatal("direct legacy path should not validate API key")
	}
}

func TestPrepareWeChatSideEffectProxyURLRequiresAPIKey(t *testing.T) {
	oldCfg, oldAccount := cfg, wechatAccountName
	oldValidate := validateAPIKeyForWeChatAccount
	t.Cleanup(func() {
		cfg = oldCfg
		wechatAccountName = oldAccount
		validateAPIKeyForWeChatAccount = oldValidate
	})

	cfg = &config.Config{
		WechatAppID:     "appid",
		WechatSecret:    "secret",
		WechatProxyURL:  "https://proxy.example.com",
		MD2WechatAPIKey: "",
	}
	wechatAccountName = ""
	validateAPIKeyForWeChatAccount = func(apiKey string) error {
		t.Fatal("validator should not be called when key is missing")
		return nil
	}

	err := prepareWeChatSideEffect()
	cliErr, ok := err.(*cliError)
	if !ok || cliErr.Code != codeAPIKeyRequired {
		t.Fatalf("prepareWeChatSideEffect() error = %#v", err)
	}
}

func TestPrepareWeChatSideEffectProxyURLUsesConfiguredAPIKey(t *testing.T) {
	oldCfg, oldAccount := cfg, wechatAccountName
	oldValidate := validateAPIKeyForWeChatAccount
	t.Cleanup(func() {
		cfg = oldCfg
		wechatAccountName = oldAccount
		validateAPIKeyForWeChatAccount = oldValidate
	})

	cfg = &config.Config{
		WechatAppID:     "appid",
		WechatSecret:    "secret",
		WechatProxyURL:  "https://proxy.example.com",
		MD2WechatAPIKey: "configured-api-key",
	}
	wechatAccountName = ""
	var gotKey string
	validateAPIKeyForWeChatAccount = func(apiKey string) error {
		gotKey = apiKey
		return nil
	}

	if err := prepareWeChatSideEffect(); err != nil {
		t.Fatalf("prepareWeChatSideEffect() error = %v", err)
	}
	if gotKey != "configured-api-key" {
		t.Fatalf("validator api key = %q, want configured-api-key", gotKey)
	}
}

func TestPrepareWeChatSideEffectProxyURLUsesAPIKeyOverride(t *testing.T) {
	oldCfg, oldAccount := cfg, wechatAccountName
	oldValidate := validateAPIKeyForWeChatAccount
	t.Cleanup(func() {
		cfg = oldCfg
		wechatAccountName = oldAccount
		validateAPIKeyForWeChatAccount = oldValidate
	})

	cfg = &config.Config{
		WechatAppID:     "appid",
		WechatSecret:    "secret",
		WechatProxyURL:  "https://proxy.example.com",
		MD2WechatAPIKey: "configured-api-key",
	}
	wechatAccountName = ""
	var gotKey string
	validateAPIKeyForWeChatAccount = func(apiKey string) error {
		gotKey = apiKey
		return nil
	}

	if err := prepareWeChatSideEffectWithAPIKey("flag-api-key"); err != nil {
		t.Fatalf("prepareWeChatSideEffectWithAPIKey() error = %v", err)
	}
	if gotKey != "flag-api-key" {
		t.Fatalf("validator api key = %q, want flag-api-key", gotKey)
	}
}

func TestPrepareWeChatSideEffectMapsInvalidAPIKey(t *testing.T) {
	oldCfg, oldAccount := cfg, wechatAccountName
	oldValidate := validateAPIKeyForWeChatAccount
	t.Cleanup(func() {
		cfg = oldCfg
		wechatAccountName = oldAccount
		validateAPIKeyForWeChatAccount = oldValidate
	})

	cfg = &config.Config{
		MD2WechatAPIKey: "bad-key",
		WechatAccounts: map[string]config.WechatAccount{
			"main": {AppID: "appid", Secret: "secret"},
		},
	}
	wechatAccountName = "main"
	validateAPIKeyForWeChatAccount = func(apiKey string) error {
		return errors.New("API_KEY_INVALID: MD2WECHAT_API_KEY is invalid")
	}

	err := prepareWeChatSideEffect()
	cliErr, ok := err.(*cliError)
	if !ok || cliErr.Code != codeAPIKeyInvalid {
		t.Fatalf("error = %#v", err)
	}
}

func TestPrepareWeChatSideEffectUsesAPIKeyOverride(t *testing.T) {
	oldCfg, oldAccount := cfg, wechatAccountName
	oldValidate := validateAPIKeyForWeChatAccount
	t.Cleanup(func() {
		cfg = oldCfg
		wechatAccountName = oldAccount
		validateAPIKeyForWeChatAccount = oldValidate
	})

	cfg = &config.Config{
		WechatAccounts: map[string]config.WechatAccount{
			"main": {AppID: "appid", Secret: "secret"},
		},
	}
	wechatAccountName = "main"
	var gotKey string
	validateAPIKeyForWeChatAccount = func(apiKey string) error {
		gotKey = apiKey
		return nil
	}

	if err := prepareWeChatSideEffectWithAPIKey("flag-api-key"); err != nil {
		t.Fatalf("prepareWeChatSideEffectWithAPIKey() error = %v", err)
	}
	if gotKey != "flag-api-key" {
		t.Fatalf("validator api key = %q, want flag-api-key", gotKey)
	}
}

func TestValidateConvertConfigFailsClosedForMissingExplicitWechatAccountWithoutSideEffect(t *testing.T) {
	oldCfg, oldAccount := cfg, wechatAccountName
	oldMode, oldAPIKey := convertMode, convertAPIKey
	oldUpload, oldDraft := convertUpload, convertDraft
	t.Cleanup(func() {
		cfg = oldCfg
		wechatAccountName = oldAccount
		convertMode = oldMode
		convertAPIKey = oldAPIKey
		convertUpload = oldUpload
		convertDraft = oldDraft
	})

	cfg = &config.Config{
		WechatAppID:        "direct-appid",
		WechatSecret:       "direct-secret",
		MD2WechatAPIKey:    "api-key",
		DefaultConvertMode: "api",
		WechatAccounts: map[string]config.WechatAccount{
			"main": {AppID: "wx-main", Secret: "secret-main"},
		},
	}
	wechatAccountName = "missing"
	convertMode = "api"
	convertAPIKey = ""
	convertUpload = false
	convertDraft = false

	err := validateConvertConfig()
	cliErr, ok := err.(*cliError)
	if !ok || cliErr.Code != codeWechatAccountNotFound {
		t.Fatalf("validateConvertConfig() error = %#v", err)
	}
}

func TestRunInspectFailsClosedForMissingExplicitWechatAccountWithoutTarget(t *testing.T) {
	oldCfg, oldAccount := cfg, wechatAccountName
	oldUpload, oldDraft := inspectUpload, inspectDraft
	t.Cleanup(func() {
		cfg = oldCfg
		wechatAccountName = oldAccount
		inspectUpload = oldUpload
		inspectDraft = oldDraft
	})

	cfg = &config.Config{
		WechatAppID:     "direct-appid",
		WechatSecret:    "direct-secret",
		MD2WechatAPIKey: "api-key",
		WechatAccounts: map[string]config.WechatAccount{
			"main": {AppID: "wx-main", Secret: "secret-main"},
		},
	}
	wechatAccountName = "missing"
	inspectDraft = false
	inspectUpload = false

	markdownPath := filepath.Join(t.TempDir(), "article.md")
	if err := os.WriteFile(markdownPath, []byte("# Title\n"), 0600); err != nil {
		t.Fatalf("write markdown: %v", err)
	}

	_, err := runInspect(markdownPath)
	cliErr, ok := err.(*cliError)
	if !ok || cliErr.Code != codeWechatAccountNotFound {
		t.Fatalf("runInspect() error = %#v", err)
	}
}

func TestRunCreateImagePostDryRunFailsClosedForMissingExplicitWechatAccount(t *testing.T) {
	oldCfg, oldAccount := cfg, wechatAccountName
	oldTitle, oldImages := imagePostTitle, imagePostImages
	oldFromMD, oldDryRun := imagePostFromMD, imagePostDryRun
	oldContent, oldOutput := imagePostContent, imagePostOutput
	t.Cleanup(func() {
		cfg = oldCfg
		wechatAccountName = oldAccount
		imagePostTitle = oldTitle
		imagePostImages = oldImages
		imagePostFromMD = oldFromMD
		imagePostDryRun = oldDryRun
		imagePostContent = oldContent
		imagePostOutput = oldOutput
	})

	cfg = &config.Config{
		WechatAppID:  "direct-appid",
		WechatSecret: "direct-secret",
		WechatAccounts: map[string]config.WechatAccount{
			"main": {AppID: "wx-main", Secret: "secret-main"},
		},
	}
	wechatAccountName = "missing"
	imagePostTitle = "Title"
	imagePostImages = "image.png"
	imagePostFromMD = ""
	imagePostDryRun = true
	imagePostContent = ""
	imagePostOutput = ""

	_, err := runCreateImagePost()
	cliErr, ok := err.(*cliError)
	if !ok || cliErr.Code != codeWechatAccountNotFound {
		t.Fatalf("runCreateImagePost() error = %#v", err)
	}
}
