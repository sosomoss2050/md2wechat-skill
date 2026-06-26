package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunTitleSuggestOutputsActionRequiredRequest(t *testing.T) {
	oldJSON := jsonOutput
	oldTargetReader := titleSuggestTargetReader
	oldCount := titleSuggestCount
	oldMaxTitleChars := titleSuggestMaxTitleChars
	oldPrompt := titleSuggestPrompt
	t.Cleanup(func() {
		jsonOutput = oldJSON
		titleSuggestTargetReader = oldTargetReader
		titleSuggestCount = oldCount
		titleSuggestMaxTitleChars = oldMaxTitleChars
		titleSuggestPrompt = oldPrompt
	})

	article := strings.Join([]string{
		"---",
		"title: Frontmatter 标题",
		"---",
		"",
		"# Body Heading",
		"",
		"这是一篇关于 Agent Native 标题工作流的正文内容。",
	}, "\n")
	articlePath := filepath.Join(t.TempDir(), "article.md")
	if err := os.WriteFile(articlePath, []byte(article), 0600); err != nil {
		t.Fatalf("write article: %v", err)
	}

	jsonOutput = true
	titleSuggestTargetReader = "AI 工具用户"
	titleSuggestCount = 10
	titleSuggestMaxTitleChars = 25
	titleSuggestPrompt = "wechat-title-expert"

	stdout := captureStdout(t, func() {
		if err := runTitleSuggest(articlePath); err != nil {
			t.Fatalf("runTitleSuggest() error = %v", err)
		}
	})

	var response map[string]any
	if err := json.Unmarshal(stdout, &response); err != nil {
		t.Fatalf("unmarshal response: %v\n%s", err, stdout)
	}
	if response["success"] != true || response["code"] != "TITLE_SUGGEST_REQUEST_READY" {
		t.Fatalf("unexpected response: %#v", response)
	}
	if response["schema_version"] != "v1" || response["status"] != "action_required" || response["retryable"] != false {
		t.Fatalf("unexpected envelope: %#v", response)
	}

	data, ok := response["data"].(map[string]any)
	if !ok {
		t.Fatalf("data type = %T", response["data"])
	}
	if data["action"] != "ai_title_suggestion_request" {
		t.Fatalf("action = %#v", data["action"])
	}
	if data["execution_owner"] != "host_agent" {
		t.Fatalf("execution_owner = %#v", data["execution_owner"])
	}
	if data["prompt_kind"] != "title" || data["prompt_name"] != "wechat-title-expert" {
		t.Fatalf("unexpected prompt identity: %#v", data)
	}
	prompt, _ := data["prompt"].(string)
	if !strings.Contains(prompt, "这是一篇关于 Agent Native 标题工作流的正文内容。") {
		t.Fatalf("prompt missing article content: %q", prompt)
	}
	if data["article_title"] != "Frontmatter 标题" {
		t.Fatalf("article_title = %#v", data["article_title"])
	}
	if data["title_count"] != float64(10) {
		t.Fatalf("title_count = %#v", data["title_count"])
	}
	if data["max_title_chars"] != float64(25) {
		t.Fatalf("max_title_chars = %#v", data["max_title_chars"])
	}
	if data["side_effects"] != false || data["requires_external_model"] != true || data["recommendation_only"] != true {
		t.Fatalf("unexpected execution flags: %#v", data)
	}
}

func TestRunTitleSuggestRequiresJSONOutput(t *testing.T) {
	oldJSON := jsonOutput
	t.Cleanup(func() { jsonOutput = oldJSON })

	jsonOutput = false

	err := runTitleSuggest("article.md")
	cliErr, ok := err.(*cliError)
	if !ok {
		t.Fatalf("error type = %T, want *cliError", err)
	}
	if cliErr.Code != codeConfigInvalid {
		t.Fatalf("code = %q, want %q", cliErr.Code, codeConfigInvalid)
	}
	if !strings.Contains(cliErr.Message, "--json") {
		t.Fatalf("message should mention --json: %q", cliErr.Message)
	}
}

func TestRunTitleSuggestMapsMissingFileToReadError(t *testing.T) {
	oldJSON := jsonOutput
	t.Cleanup(func() { jsonOutput = oldJSON })

	jsonOutput = true

	err := runTitleSuggest(filepath.Join(t.TempDir(), "missing.md"))
	cliErr, ok := err.(*cliError)
	if !ok {
		t.Fatalf("error type = %T, want *cliError", err)
	}
	if cliErr.Code != "TITLE_SUGGEST_READ_FAILED" {
		t.Fatalf("code = %q", cliErr.Code)
	}
}

func TestRunTitleSuggestMapsInvalidCountToInvalidError(t *testing.T) {
	oldJSON := jsonOutput
	oldCount := titleSuggestCount
	t.Cleanup(func() {
		jsonOutput = oldJSON
		titleSuggestCount = oldCount
	})

	articlePath := filepath.Join(t.TempDir(), "article.md")
	if err := os.WriteFile(articlePath, []byte("# Title\n\n正文内容。"), 0600); err != nil {
		t.Fatalf("write article: %v", err)
	}

	jsonOutput = true
	titleSuggestCount = 7

	err := runTitleSuggest(articlePath)
	cliErr, ok := err.(*cliError)
	if !ok {
		t.Fatalf("error type = %T, want *cliError", err)
	}
	if cliErr.Code != "TITLE_SUGGEST_INVALID" {
		t.Fatalf("code = %q", cliErr.Code)
	}
}
