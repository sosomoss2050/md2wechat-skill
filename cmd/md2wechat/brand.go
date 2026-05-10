package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/geekjourneyx/md2wechat-skill/internal/action"
	"github.com/spf13/cobra"
)

// brandCmd brand 命令
var brandCmd = &cobra.Command{
	Use:   "brand",
	Short: "Manage Brand Profile for AI agents",
	Long: `Manage Brand Profile for AI agents.

The Brand Profile is a Markdown file that AI agents read to understand your
voice, layout preferences, and constraints when generating content.

This file is stored at ~/.config/md2wechat/brand.md and is NOT parsed
by the CLI itself. It is purely for agent consumption.

Documentation: docs/BRAND-PROFILE.md`,
}

func init() {
	// init 子命令
	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Create a Brand Profile template",
		Long: `Create a Brand Profile template at ~/.config/md2wechat/brand.md.

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
		Long:  `Show the content of the Brand Profile at ~/.config/md2wechat/brand.md.`,
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
	return filepath.Join(homeDir, ".config", "md2wechat", "brand.md"), nil
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
	createdDate := time.Now().Format("2006-01-02")
	template := fmt.Sprintf(`# 品牌档案 / Brand Profile

> 此文件由 AI Agent 读取，CLI 不解析。请用自然语言描述你的创作风格。
> 配置指南：docs/BRAND-PROFILE.md | 命令帮助：md2wechat brand --help

---

## 基本信息

**名字 / 品牌名**：

**简介**：（一句话介绍你是谁、写什么）

---

## 语气与风格

描述你希望文章呈现的语气。可以写具体例子和反例，越具体越好。

**我的风格**：

**我要避免的表达**：
- （例如：过多 emoji，空泛鸡汤，过度营销词汇）

---

## 文章开头偏好

参考选项：verdict_first（先结论）/ story_first（先故事）/ question_first（先问题）/ data_first（先数据）

**我的偏好**：

---

## 排版约束

Agent 会遵守以下数量约束（可修改数字）：

- 最多模块数：6（上限 43，填 0 使用默认值）
- 最多 CTA 数：1（上限 2）
- 最多引用数：2（上限 10）
- 最多 Hero 数：1（固定上限）

---

## 默认 CTA（行动引导）

**标题**：（例如：如果这篇对你有启发）
**正文**：（例如：欢迎关注，我在持续记录 AI 工具和独立开发实践）
**行动**：（例如：关注 / 咨询 / 分享）

---

## 作者卡片

**名字**：
**头衔**：（例如：AI 应用开发者 / 独立开发者）
**简介**：（2-3 句话，介绍你的背景和关注领域）

---

## 风格参考文件（可选）

如果你有更详细的写作风格指南文件或目录，在此填写路径（Agent 会读取全文）：

**路径**：（例如：~/Documents/brand/voice-guide.md）

---

*创建时间：%s*
`, createdDate)

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
			Message:       "Brand Profile not found. Run 'md2wechat brand init' to create one.",
			SchemaVersion: action.SchemaVersion,
			Status:        action.StatusActionRequired,
			Retryable:     false,
		})
		return nil
	}

	// 读取文件（Markdown 无需解析，直接返回原始内容）
	content, err := os.ReadFile(brandPath)
	if err != nil {
		return wrapCLIError(codeBrandReadFailed, err, fmt.Sprintf("failed to read file: %s", brandPath))
	}

	displayPath := normalizeBrandPath(brandPath)
	responseSuccessWith(codeBrandShown, "Brand Profile loaded successfully", map[string]any{
		"path":    displayPath,
		"content": string(content),
	})
	return nil
}
