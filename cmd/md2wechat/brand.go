package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/geekjourneyx/md2wechat-skill/internal/action"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// brandCmd brand 命令
var brandCmd = &cobra.Command{
	Use:   "brand",
	Short: "Manage Brand Profile for AI agents",
	Long: `Manage Brand Profile for AI agents.

The Brand Profile is a YAML file that AI agents read to understand your
voice, layout preferences, and constraints when generating content.

This file is stored at ~/.config/md2wechat/brand.yaml and is NOT parsed
by the CLI itself. It is purely for agent consumption.

Documentation: docs/AGENT-GUIDE.md`,
}

func init() {
	// init 子命令
	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Create a Brand Profile template",
		Long: `Create a Brand Profile template at ~/.config/md2wechat/brand.yaml.

If the file already exists, this command is idempotent and will not overwrite it.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBrandInit()
		},
	}
	brandCmd.AddCommand(initCmd)

	// show 子命令
	var showCmd = &cobra.Command{
		Use:   "show",
		Short: "Show current Brand Profile",
		Long:  `Show the content of the Brand Profile at ~/.config/md2wechat/brand.yaml.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBrandShow()
		},
	}
	brandCmd.AddCommand(showCmd)
}

// getBrandProfilePath 获取 Brand Profile 文件路径
func getBrandProfilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "md2wechat", "brand.yaml"), nil
}

// normalizeBrandPath 将路径转换为 ~/... 格式
func normalizeBrandPath(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, homeDir) {
		rel := strings.TrimPrefix(path, homeDir)
		if strings.HasPrefix(rel, "/") || strings.HasPrefix(rel, "\\") {
			rel = rel[1:]
		}
		return "~/" + rel
	}
	return path
}

// runBrandInit 初始化 Brand Profile
func runBrandInit() error {
	brandPath, err := getBrandProfilePath()
	if err != nil {
		return wrapCLIError(codeBrandInitFailed, err, "failed to determine home directory")
	}

	// 检查文件是否已存在（幂等性）
	if _, err := os.Stat(brandPath); err == nil {
		// 文件已存在，返回成功（幂等）
		displayPath := normalizeBrandPath(brandPath)
		responseSuccessWith(codeBrandInitialized, "Brand Profile already exists", map[string]any{
			"file":    displayPath,
			"message": "Brand Profile already exists (not overwritten)",
		})
		return nil
	}

	// 创建目录（如果不存在）
	brandDir := filepath.Dir(brandPath)
	if err := os.MkdirAll(brandDir, 0755); err != nil {
		return wrapCLIError(codeBrandInitFailed, err, fmt.Sprintf("failed to create directory: %s", brandDir))
	}

	// 创建模板内容
	template := `# md2wechat Brand Profile
# 此文件由 Agent 读取，CLI 不解析此文件
# 文档：docs/AGENT-GUIDE.md
schema_version: 1

# 你的名字或品牌名
name: ""

voice:
  # 语气风格描述（例如：犀利实用，第一人称）
  tone: ""
  # 可选：风格参考文件或目录的绝对路径
  # style_ref: ~/Documents/brand/voice-guide.md
  # 要避免的表达方式
  avoid: []

layout:
  # 文章开头风格：verdict_first | story_first | question_first | data_first
  opening: ""

limits:
  max_modules: 6   # 上限 43，设为 0 则使用默认值 6
  max_cta: 1       # 上限 2
  max_quotes: 2    # 上限 10
  max_hero: 1      # 上限 1

cta:
  default_title: ""
  default_body: ""
  default_action: ""

author_card:
  name: ""
  title: ""
  bio: ""
`

	// 写入文件
	if err := os.WriteFile(brandPath, []byte(template), 0644); err != nil {
		return wrapCLIError(codeBrandInitFailed, err, fmt.Sprintf("failed to write file: %s", brandPath))
	}

	displayPath := normalizeBrandPath(brandPath)
	if !jsonOutput {
		fmt.Fprintf(os.Stderr, "\n✅ Brand Profile 已创建: %s\n", displayPath)
		fmt.Fprintf(os.Stderr, "📝 下一步: 编辑此文件，填入你的品牌信息和风格偏好\n")
		fmt.Fprintf(os.Stderr, "📍 文档: docs/AGENT-GUIDE.md\n\n")
	}

	responseSuccessWith(codeBrandInitialized, "Brand Profile created successfully", map[string]any{
		"file":    displayPath,
		"message": "Brand Profile created successfully. Please edit it with your brand information.",
	})
	return nil
}

// runBrandShow 显示 Brand Profile
func runBrandShow() error {
	brandPath, err := getBrandProfilePath()
	if err != nil {
		return wrapCLIError(codeBrandReadFailed, err, "failed to determine home directory")
	}

	// 检查文件是否存在
	if _, err := os.Stat(brandPath); os.IsNotExist(err) {
		responseWith(cliResponse{
			Success:       false,
			Code:          codeBrandNotFound,
			Message:       fmt.Sprintf("Brand Profile not found. Run 'md2wechat brand init' to create one."),
			SchemaVersion: action.SchemaVersion,
			Status:        action.StatusActionRequired,
			Retryable:     false,
		})
		return nil
	}

	// 读取文件
	content, err := os.ReadFile(brandPath)
	if err != nil {
		return wrapCLIError(codeBrandReadFailed, err, fmt.Sprintf("failed to read file: %s", brandPath))
	}

	// 解析 YAML
	var profile map[string]any
	if err := yaml.Unmarshal(content, &profile); err != nil {
		responseWith(cliResponse{
			Success:       false,
			Code:          codeBrandReadFailed,
			Message:       "Failed to parse Brand Profile (invalid YAML)",
			SchemaVersion: action.SchemaVersion,
			Status:        action.StatusFailed,
			Retryable:     false,
			Error:         err.Error(),
		})
		return nil
	}

	displayPath := normalizeBrandPath(brandPath)
	responseSuccessWith(codeBrandShown, "Brand Profile loaded successfully", map[string]any{
		"path":    displayPath,
		"profile": profile,
	})
	return nil
}
