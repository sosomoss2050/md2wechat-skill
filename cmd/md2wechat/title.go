package main

import (
	"fmt"
	"os"

	"github.com/geekjourneyx/md2wechat-skill/internal/converter"
	titlebuilder "github.com/geekjourneyx/md2wechat-skill/internal/title"
	"github.com/spf13/cobra"
)

var (
	titleSuggestTargetReader  string
	titleSuggestCount         int
	titleSuggestMaxTitleChars int
	titleSuggestPrompt        string
)

var titleCmd = &cobra.Command{
	Use:   "title",
	Short: "Prepare title optimization workflows",
}

var titleSuggestCmd = &cobra.Command{
	Use:   "suggest <article.md>",
	Short: "Prepare an AI request for WeChat title suggestions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTitleSuggest(args[0])
	},
}

type titleSuggestData struct {
	titlebuilder.SuggestAIRequest
	ArticlePath string `json:"article_path,omitempty"`
}

func init() {
	titleSuggestCmd.Flags().StringVar(&titleSuggestTargetReader, "target-reader", "", "Target reader for title suggestions")
	titleSuggestCmd.Flags().IntVar(&titleSuggestCount, "count", titlebuilder.DefaultCount, "Number of title candidates to request")
	titleSuggestCmd.Flags().IntVar(&titleSuggestMaxTitleChars, "max-title-chars", titlebuilder.DefaultMaxTitleChars, "Maximum characters per title")
	titleSuggestCmd.Flags().StringVar(&titleSuggestPrompt, "prompt", titlebuilder.DefaultPromptName, "Title prompt preset name")
	titleCmd.AddCommand(titleSuggestCmd)
}

func runTitleSuggest(articlePath string) error {
	if !jsonOutput {
		return newCLIError(codeConfigInvalid, "title suggest requires --json for machine-readable Agent Native output")
	}

	markdown, err := os.ReadFile(articlePath)
	if err != nil {
		return wrapCLIError(codeTitleSuggestReadFailed, err, fmt.Sprintf("read article for title suggest: %v", err))
	}

	doc := converter.ParseArticleDocument(string(markdown))
	request, err := titlebuilder.BuildSuggestRequest(titlebuilder.SuggestRequest{
		ArticleContent: doc.Body,
		ExistingTitle:  doc.Metadata.Title,
		TargetReader:   titleSuggestTargetReader,
		Count:          titleSuggestCount,
		MaxTitleChars:  titleSuggestMaxTitleChars,
		PromptName:     titleSuggestPrompt,
	})
	if err != nil {
		return wrapCLIError(codeTitleSuggestInvalid, err, err.Error())
	}

	responseActionRequiredWith(codeTitleSuggestRequestReady, "Title suggestion AI request prepared", titleSuggestData{
		SuggestAIRequest: *request,
		ArticlePath:      articlePath,
	})
	return nil
}
