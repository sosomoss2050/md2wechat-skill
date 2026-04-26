package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/geekjourneyx/md2wechat-skill/internal/layoutcatalog"
)

func TestLayoutListJSONIncludesHero(t *testing.T) {
	oldJSON := jsonOutput
	t.Cleanup(func() {
		jsonOutput = oldJSON
		layoutcatalog.ResetDefaultCatalogForTests()
	})

	jsonOutput = true
	layoutcatalog.ResetDefaultCatalogForTests()

	stdout := captureStdout(t, func() {
		if err := layoutListCmd.RunE(layoutListCmd, nil); err != nil {
			t.Fatalf("layoutListCmd.RunE() error = %v", err)
		}
	})

	if !strings.Contains(string(stdout), `"hero"`) {
		t.Errorf("expected hero in list output, got:\n%s", stdout)
	}
}

func TestLayoutShowJSONReturnsSpec(t *testing.T) {
	oldJSON := jsonOutput
	t.Cleanup(func() {
		jsonOutput = oldJSON
		layoutcatalog.ResetDefaultCatalogForTests()
	})

	jsonOutput = true
	layoutcatalog.ResetDefaultCatalogForTests()

	stdout := captureStdout(t, func() {
		if err := layoutShowCmd.RunE(layoutShowCmd, []string{"hero"}); err != nil {
			t.Fatalf("layoutShowCmd.RunE() error = %v", err)
		}
	})

	var response map[string]any
	if err := json.Unmarshal(stdout, &response); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout)
	}
	if response["success"] != true {
		t.Fatalf("expected success in response: %#v", response)
	}
	data, _ := response["data"].(map[string]any)
	if data["spec"] == nil {
		t.Fatalf("expected spec in response data: %#v", data)
	}
}

func TestLayoutRenderHeroProducesBlock(t *testing.T) {
	oldJSON := jsonOutput
	oldVars := append([]string(nil), layoutRenderVars...)
	t.Cleanup(func() {
		jsonOutput = oldJSON
		layoutRenderVars = oldVars
		layoutcatalog.ResetDefaultCatalogForTests()
	})

	jsonOutput = true
	layoutcatalog.ResetDefaultCatalogForTests()
	layoutRenderVars = []string{"eyebrow=深度观察", "title=公众号排版的真问题"}

	stdout := captureStdout(t, func() {
		if err := layoutRenderCmd.RunE(layoutRenderCmd, []string{"hero"}); err != nil {
			t.Fatalf("layoutRenderCmd.RunE() error = %v", err)
		}
	})

	if !strings.Contains(string(stdout), `:::hero`) {
		t.Errorf("expected :::hero in output:\n%s", stdout)
	}
}

func TestLayoutValidateUnknownWarns(t *testing.T) {
	oldJSON := jsonOutput
	oldStdin := layoutValidateStdin
	oldReader := stdinReader
	t.Cleanup(func() {
		jsonOutput = oldJSON
		layoutValidateStdin = oldStdin
		stdinReader = oldReader
		layoutcatalog.ResetDefaultCatalogForTests()
	})

	jsonOutput = true
	layoutValidateStdin = true
	stdinReader = strings.NewReader(":::futuristic-block\nfoo: bar\n:::\n")
	layoutcatalog.ResetDefaultCatalogForTests()

	stdout := captureStdout(t, func() {
		// unknown block produces a warning, not an error — result is informational
		_ = layoutValidateCmd.RunE(layoutValidateCmd, nil)
	})

	if !strings.Contains(string(stdout), "futuristic-block") {
		t.Errorf("expected unknown module to appear in warnings:\n%s", stdout)
	}
}

func TestLayoutShowNotFound(t *testing.T) {
	oldJSON := jsonOutput
	t.Cleanup(func() {
		jsonOutput = oldJSON
		layoutcatalog.ResetDefaultCatalogForTests()
	})

	jsonOutput = true
	layoutcatalog.ResetDefaultCatalogForTests()

	if err := layoutShowCmd.RunE(layoutShowCmd, []string{"nonexistent-module-xyz"}); err == nil {
		t.Fatal("expected error for nonexistent module")
	} else if cliErr, ok := err.(*cliError); !ok || cliErr.Code != codeLayoutModuleNotFound {
		t.Fatalf("unexpected error: %#v", err)
	}
}

func TestRenderCmdRowsJSONInput(t *testing.T) {
	oldJSON := jsonOutput
	oldVars := append([]string(nil), layoutRenderVars...)
	t.Cleanup(func() {
		jsonOutput = oldJSON
		layoutRenderVars = oldVars
		layoutcatalog.ResetDefaultCatalogForTests()
	})

	jsonOutput = true
	layoutcatalog.ResetDefaultCatalogForTests()
	// toc rows schema: number | title | description (min_columns: 2)
	layoutRenderVars = []string{`rows=[["01","第一章","概述"]]`}

	stdout := captureStdout(t, func() {
		if err := layoutRenderCmd.RunE(layoutRenderCmd, []string{"toc"}); err != nil {
			t.Fatalf("layoutRenderCmd.RunE() error = %v", err)
		}
	})

	if !strings.Contains(string(stdout), ":::toc") {
		t.Errorf("expected :::toc in output:\n%s", stdout)
	}
}

func TestLayoutRenderMissingRequiredField(t *testing.T) {
	oldJSON := jsonOutput
	oldVars := append([]string(nil), layoutRenderVars...)
	t.Cleanup(func() {
		jsonOutput = oldJSON
		layoutRenderVars = oldVars
		layoutcatalog.ResetDefaultCatalogForTests()
	})

	jsonOutput = true
	layoutcatalog.ResetDefaultCatalogForTests()
	// hero requires eyebrow and title — omit both
	layoutRenderVars = nil

	if err := layoutRenderCmd.RunE(layoutRenderCmd, []string{"hero"}); err == nil {
		t.Fatal("expected error for missing required field")
	} else if cliErr, ok := err.(*cliError); !ok || cliErr.Code != codeLayoutMissingRequiredField {
		t.Fatalf("unexpected error code: %#v", err)
	}
}
