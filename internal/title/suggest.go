package title

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/geekjourneyx/md2wechat-skill/internal/promptcatalog"
)

const (
	DefaultPromptName = "wechat-title-expert"
	PromptKind        = "title"

	DefaultCount = 10
	MinCount     = 8
	MaxCount     = 10

	DefaultHookLevel = 1
	MinHookLevel     = 1
	MaxHookLevel     = 3

	DefaultMaxTitleChars  = 25
	MinTitleChars         = 12
	MetadataTitleMaxChars = 32
)

const (
	actionTitleSuggestionRequest = "ai_title_suggestion_request"
	executionOwnerHostAgent      = "host_agent"
	neutralTargetReader          = "未指定，请从文章内容推断"
)

type SuggestRequest struct {
	ArticleContent string
	ExistingTitle  string
	TargetReader   string
	Count          int
	HookLevel      int
	MaxTitleChars  int
	PromptName     string
}

type SuggestAIRequest struct {
	Action                string `json:"action"`
	ExecutionOwner        string `json:"execution_owner"`
	PromptKind            string `json:"prompt_kind"`
	PromptName            string `json:"prompt_name"`
	Prompt                string `json:"prompt"`
	ArticleTitle          string `json:"article_title,omitempty"`
	ArticleChars          int    `json:"article_chars"`
	TargetReader          string `json:"target_reader"`
	TitleCount            int    `json:"title_count"`
	MaxTitleChars         int    `json:"max_title_chars"`
	HookLevel             int    `json:"hook_level"`
	HookLevelLabel        string `json:"hook_level_label"`
	SideEffects           bool   `json:"side_effects"`
	RequiresExternalModel bool   `json:"requires_external_model"`
	RecommendationOnly    bool   `json:"recommendation_only"`
}

func BuildSuggestRequest(req SuggestRequest) (*SuggestAIRequest, error) {
	articleContent := strings.TrimSpace(req.ArticleContent)
	if articleContent == "" {
		return nil, fmt.Errorf("article content is required")
	}

	count := req.Count
	if count == 0 {
		count = DefaultCount
	}
	if count < MinCount || count > MaxCount {
		return nil, fmt.Errorf("count must be between %d and %d: %d", MinCount, MaxCount, count)
	}

	maxTitleChars := req.MaxTitleChars
	if maxTitleChars == 0 {
		maxTitleChars = DefaultMaxTitleChars
	}
	if maxTitleChars < MinTitleChars || maxTitleChars > MetadataTitleMaxChars {
		return nil, fmt.Errorf("max title chars must be between %d and %d: %d", MinTitleChars, MetadataTitleMaxChars, maxTitleChars)
	}

	hookLevel := req.HookLevel
	if hookLevel == 0 {
		hookLevel = DefaultHookLevel
	}
	if hookLevel < MinHookLevel || hookLevel > MaxHookLevel {
		return nil, fmt.Errorf("hook level must be between %d and %d: %d", MinHookLevel, MaxHookLevel, hookLevel)
	}
	hookLevelLabel := HookLevelLabel(hookLevel)

	promptName := strings.TrimSpace(req.PromptName)
	if promptName == "" {
		promptName = DefaultPromptName
	}

	targetReader := strings.TrimSpace(req.TargetReader)
	if targetReader == "" {
		targetReader = neutralTargetReader
	}

	prompt, err := renderBundledPrompt(promptName, map[string]string{
		"ARTICLE_CONTENT":  articleContent,
		"TARGET_READER":    targetReader,
		"TITLE_COUNT":      strconv.Itoa(count),
		"MAX_TITLE_CHARS":  strconv.Itoa(maxTitleChars),
		"HOOK_LEVEL":       strconv.Itoa(hookLevel),
		"HOOK_LEVEL_LABEL": hookLevelLabel,
	})
	if err != nil {
		return nil, fmt.Errorf("render title suggestion prompt %s/%s: %w", PromptKind, promptName, err)
	}

	return &SuggestAIRequest{
		Action:                actionTitleSuggestionRequest,
		ExecutionOwner:        executionOwnerHostAgent,
		PromptKind:            PromptKind,
		PromptName:            promptName,
		Prompt:                prompt,
		ArticleTitle:          strings.TrimSpace(req.ExistingTitle),
		ArticleChars:          len([]rune(articleContent)),
		TargetReader:          targetReader,
		TitleCount:            count,
		MaxTitleChars:         maxTitleChars,
		HookLevel:             hookLevel,
		HookLevelLabel:        hookLevelLabel,
		SideEffects:           false,
		RequiresExternalModel: true,
		RecommendationOnly:    true,
	}, nil
}

func HookLevelLabel(level int) string {
	switch level {
	case 1:
		return "restrained"
	case 2:
		return "punchy"
	case 3:
		return "high_tension"
	default:
		return ""
	}
}

func renderBundledPrompt(name string, vars map[string]string) (string, error) {
	cat, err := promptcatalog.DefaultCatalog()
	if err != nil {
		return "", fmt.Errorf("load prompt catalog: %w", err)
	}
	rendered, spec, err := cat.Render(PromptKind, name, vars)
	if err != nil {
		return "", err
	}
	if spec.Kind != PromptKind || spec.Name != name {
		return "", fmt.Errorf("prompt identity mismatch: got %s/%s", spec.Kind, spec.Name)
	}
	if strings.TrimSpace(rendered) == "" {
		return "", fmt.Errorf("rendered prompt is empty")
	}
	return rendered, nil
}
