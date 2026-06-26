package title

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/geekjourneyx/md2wechat-skill/internal/assets"
	"gopkg.in/yaml.v3"
)

const (
	DefaultPromptName = "wechat-title-expert"
	PromptKind        = "title"

	DefaultCount = 10
	MinCount     = 8
	MaxCount     = 10

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
	MaxTitleChars  int
	PromptName     string
}

type SuggestAIRequest struct {
	Action                string
	ExecutionOwner        string
	PromptKind            string
	PromptName            string
	Prompt                string
	ArticleTitle          string
	ArticleChars          int
	TargetReader          string
	TitleCount            int
	MaxTitleChars         int
	SideEffects           bool
	RequiresExternalModel bool
	RecommendationOnly    bool
}

type promptSpec struct {
	Name     string `yaml:"name"`
	Kind     string `yaml:"kind"`
	Template string `yaml:"template"`
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

	promptName := strings.TrimSpace(req.PromptName)
	if promptName == "" {
		promptName = DefaultPromptName
	}

	targetReader := strings.TrimSpace(req.TargetReader)
	if targetReader == "" {
		targetReader = neutralTargetReader
	}

	prompt, err := renderBundledPrompt(promptName, map[string]string{
		"ARTICLE_CONTENT": articleContent,
		"TARGET_READER":   targetReader,
		"TITLE_COUNT":     strconv.Itoa(count),
		"MAX_TITLE_CHARS": strconv.Itoa(maxTitleChars),
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
		SideEffects:           false,
		RequiresExternalModel: true,
		RecommendationOnly:    true,
	}, nil
}

func renderBundledPrompt(name string, vars map[string]string) (string, error) {
	if name != DefaultPromptName {
		return "", fmt.Errorf("prompt not found: %s/%s", PromptKind, name)
	}

	data, err := assets.ReadBuiltinPrompt(PromptKind, name)
	if err != nil {
		return "", fmt.Errorf("read bundled prompt: %w", err)
	}

	var spec promptSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return "", fmt.Errorf("parse bundled prompt: %w", err)
	}
	if spec.Kind != PromptKind || spec.Name != name {
		return "", fmt.Errorf("bundled prompt identity mismatch: got %s/%s", spec.Kind, spec.Name)
	}
	if strings.TrimSpace(spec.Template) == "" {
		return "", fmt.Errorf("bundled prompt template is empty")
	}

	rendered := spec.Template
	for key, value := range vars {
		rendered = strings.ReplaceAll(rendered, "{{"+key+"}}", value)
	}
	return rendered, nil
}
