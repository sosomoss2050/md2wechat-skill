package title

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/geekjourneyx/md2wechat-skill/internal/promptcatalog"
)

func TestBuildSuggestRequestDefaultsRenderBundledPrompt(t *testing.T) {
	promptcatalog.ResetDefaultCatalogForTests()
	t.Chdir(t.TempDir())

	article := "这篇文章复盘了一个公众号标题实验，重点是用真实价值降低标题党风险。"

	got, err := BuildSuggestRequest(SuggestRequest{
		ArticleContent: "  " + article + "  ",
		ExistingTitle:  "旧标题",
	})
	if err != nil {
		t.Fatalf("BuildSuggestRequest() error = %v", err)
	}

	if got.Action != "ai_title_suggestion_request" {
		t.Fatalf("Action = %q", got.Action)
	}
	if got.ExecutionOwner != "host_agent" {
		t.Fatalf("ExecutionOwner = %q", got.ExecutionOwner)
	}
	if got.PromptKind != PromptKind || got.PromptName != DefaultPromptName {
		t.Fatalf("prompt identity = %q/%q", got.PromptKind, got.PromptName)
	}
	if got.ArticleTitle != "旧标题" {
		t.Fatalf("ArticleTitle = %q", got.ArticleTitle)
	}
	if got.ArticleChars != len([]rune(article)) {
		t.Fatalf("ArticleChars = %d, want %d", got.ArticleChars, len([]rune(article)))
	}
	if got.TargetReader != neutralTargetReader {
		t.Fatalf("TargetReader = %q", got.TargetReader)
	}
	if got.TitleCount != DefaultCount {
		t.Fatalf("TitleCount = %d", got.TitleCount)
	}
	if got.MaxTitleChars != DefaultMaxTitleChars {
		t.Fatalf("MaxTitleChars = %d", got.MaxTitleChars)
	}
	if got.SideEffects {
		t.Fatal("SideEffects = true")
	}
	if !got.RequiresExternalModel {
		t.Fatal("RequiresExternalModel = false")
	}
	if !got.RecommendationOnly {
		t.Fatal("RecommendationOnly = false")
	}
	if !strings.Contains(got.Prompt, article) {
		t.Fatalf("Prompt missing article content: %q", got.Prompt)
	}
	if !strings.Contains(got.Prompt, "候选数量：10") || !strings.Contains(got.Prompt, `"title_count": 10`) {
		t.Fatalf("Prompt missing default title count: %q", got.Prompt)
	}
	if !strings.Contains(got.Prompt, "标题最长字数：25") || !strings.Contains(got.Prompt, `"max_title_chars": 25`) {
		t.Fatalf("Prompt missing default max title chars: %q", got.Prompt)
	}
}

func TestBuildSuggestRequestUsesStableJSONFieldNames(t *testing.T) {
	promptcatalog.ResetDefaultCatalogForTests()
	t.Chdir(t.TempDir())

	got, err := BuildSuggestRequest(SuggestRequest{
		ArticleContent: "有效文章内容",
		ExistingTitle:  "旧标题",
	})
	if err != nil {
		t.Fatalf("BuildSuggestRequest() error = %v", err)
	}

	data, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	payload := string(data)
	for _, want := range []string{
		`"execution_owner"`,
		`"prompt_kind"`,
		`"prompt_name"`,
		`"article_title"`,
		`"article_chars"`,
		`"title_count"`,
		`"max_title_chars"`,
		`"side_effects"`,
		`"requires_external_model"`,
		`"recommendation_only"`,
	} {
		if !strings.Contains(payload, want) {
			t.Fatalf("JSON missing %s: %s", want, payload)
		}
	}
	for _, unwanted := range []string{"ExecutionOwner", "PromptKind", "ArticleTitle", "RequiresExternalModel"} {
		if strings.Contains(payload, unwanted) {
			t.Fatalf("JSON leaked Go field name %q: %s", unwanted, payload)
		}
	}
}

func TestBuildSuggestRequestRejectsEmptyArticleContent(t *testing.T) {
	_, err := BuildSuggestRequest(SuggestRequest{ArticleContent: " \n\t "})
	if err == nil {
		t.Fatal("BuildSuggestRequest() error = nil")
	}
	if !strings.Contains(err.Error(), "article content is required") {
		t.Fatalf("error = %v", err)
	}
}

func TestBuildSuggestRequestRejectsCountOutsideRange(t *testing.T) {
	promptcatalog.ResetDefaultCatalogForTests()
	t.Chdir(t.TempDir())

	for _, tc := range []struct {
		name  string
		count int
	}{
		{name: "below min", count: 7},
		{name: "above max", count: 11},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := BuildSuggestRequest(SuggestRequest{
				ArticleContent: "有效文章内容",
				Count:          tc.count,
			})
			if err == nil {
				t.Fatal("BuildSuggestRequest() error = nil")
			}
			if !strings.Contains(err.Error(), "count must be between 8 and 10") {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestBuildSuggestRequestRejectsMaxTitleCharsOutsideRange(t *testing.T) {
	promptcatalog.ResetDefaultCatalogForTests()
	t.Chdir(t.TempDir())

	for _, tc := range []struct {
		name          string
		maxTitleChars int
	}{
		{name: "below min", maxTitleChars: 11},
		{name: "above metadata title ceiling", maxTitleChars: 33},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := BuildSuggestRequest(SuggestRequest{
				ArticleContent: "有效文章内容",
				MaxTitleChars:  tc.maxTitleChars,
			})
			if err == nil {
				t.Fatal("BuildSuggestRequest() error = nil")
			}
			if !strings.Contains(err.Error(), "max title chars must be between 12 and 32") {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestBuildSuggestRequestCustomInputsAreRenderedAndStructured(t *testing.T) {
	promptcatalog.ResetDefaultCatalogForTests()
	t.Chdir(t.TempDir())

	got, err := BuildSuggestRequest(SuggestRequest{
		ArticleContent: "本文讲给独立开发者看的增长复盘，强调小样本实验和标题承诺一致。",
		ExistingTitle:  "原始增长复盘",
		TargetReader:   "独立开发者",
		Count:          8,
		MaxTitleChars:  22,
	})
	if err != nil {
		t.Fatalf("BuildSuggestRequest() error = %v", err)
	}

	if got.ArticleTitle != "原始增长复盘" {
		t.Fatalf("ArticleTitle = %q", got.ArticleTitle)
	}
	if got.TargetReader != "独立开发者" {
		t.Fatalf("TargetReader = %q", got.TargetReader)
	}
	if got.TitleCount != 8 {
		t.Fatalf("TitleCount = %d", got.TitleCount)
	}
	if got.MaxTitleChars != 22 {
		t.Fatalf("MaxTitleChars = %d", got.MaxTitleChars)
	}
	for _, want := range []string{
		"目标读者：独立开发者",
		"候选数量：8",
		"标题最长字数：22",
		`"target_reader": "独立开发者"`,
		`"title_count": 8`,
		`"max_title_chars": 22`,
	} {
		if !strings.Contains(got.Prompt, want) {
			t.Fatalf("Prompt missing %q: %q", want, got.Prompt)
		}
	}
}

func TestBuildSuggestRequestInvalidPromptNameReturnsUsefulError(t *testing.T) {
	promptcatalog.ResetDefaultCatalogForTests()
	t.Chdir(t.TempDir())

	_, err := BuildSuggestRequest(SuggestRequest{
		ArticleContent: "有效文章内容",
		PromptName:     "missing-title-prompt",
	})
	if err == nil {
		t.Fatal("BuildSuggestRequest() error = nil")
	}
	if !strings.Contains(err.Error(), "render title suggestion prompt") ||
		!strings.Contains(err.Error(), "prompt not found: title/missing-title-prompt") {
		t.Fatalf("error = %v", err)
	}
}
