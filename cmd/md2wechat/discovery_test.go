package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/geekjourneyx/md2wechat-skill/internal/promptcatalog"
)

func TestBuildProviderViewsIncludesBuiltinProviders(t *testing.T) {
	oldCfg := cfg
	t.Cleanup(func() { cfg = oldCfg })

	cfg = nil
	providers, err := buildProviderViews()
	if err != nil {
		t.Fatalf("buildProviderViews() error = %v", err)
	}
	if len(providers) == 0 {
		t.Fatal("expected providers")
	}
	found := false
	for _, provider := range providers {
		if provider.Name == "openai" {
			found = true
			if !provider.SupportsSize {
				t.Fatalf("expected openai SupportsSize")
			}
		}
	}
	if !found {
		t.Fatal("expected openai provider")
	}
}

func TestListThemesIncludesBuiltinTheme(t *testing.T) {
	themes, err := listThemes()
	if err != nil {
		t.Fatalf("listThemes() error = %v", err)
	}
	found := false
	for _, theme := range themes {
		if theme.Name == "default" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected builtin default theme")
	}
}

func TestBuildCapabilitiesDataIncludesPromptCatalog(t *testing.T) {
	oldCfg := cfg
	t.Cleanup(func() {
		cfg = oldCfg
		promptcatalog.ResetDefaultCatalogForTests()
	})

	cfg = nil
	promptcatalog.ResetDefaultCatalogForTests()

	data, err := buildCapabilitiesData()
	if err != nil {
		t.Fatalf("buildCapabilitiesData() error = %v", err)
	}
	prompts, ok := data["prompts"].([]promptcatalog.PromptSpec)
	if !ok || len(prompts) == 0 {
		t.Fatalf("expected prompt catalog in capabilities: %#v", data["prompts"])
	}
}

func TestPromptsRenderCommandUsesStableJSONEnvelope(t *testing.T) {
	oldJSON := jsonOutput
	oldPromptKind := promptKind
	oldPromptVars := append([]string(nil), promptVars...)
	t.Cleanup(func() {
		jsonOutput = oldJSON
		promptKind = oldPromptKind
		promptVars = oldPromptVars
		promptcatalog.ResetDefaultCatalogForTests()
	})

	jsonOutput = true
	promptcatalog.ResetDefaultCatalogForTests()
	promptKind = "image"
	promptVars = []string{"ARTICLE_TITLE=测试标题", "ARTICLE_SUMMARY=测试摘要", "VISUAL_STYLE=极简"}

	stdout := captureStdout(t, func() {
		if err := promptsRenderCmd.RunE(promptsRenderCmd, []string{"cover-default"}); err != nil {
			t.Fatalf("RunE() error = %v", err)
		}
	})

	var response map[string]any
	if err := json.Unmarshal(stdout, &response); err != nil {
		t.Fatalf("unmarshal response: %v\n%s", err, stdout)
	}
	if response["success"] != true || response["code"] != "PROMPTS_SHOWN" {
		t.Fatalf("unexpected response: %#v", response)
	}
	data, _ := response["data"].(map[string]any)
	rendered, _ := data["rendered"].(string)
	if !strings.Contains(rendered, "测试标题") {
		t.Fatalf("rendered = %q", rendered)
	}
}
