package promptcatalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultCatalogLoadsBuiltinPrompts(t *testing.T) {
	ResetDefaultCatalogForTests()
	t.Chdir(t.TempDir())

	cat, err := DefaultCatalog()
	if err != nil {
		t.Fatalf("DefaultCatalog() error = %v", err)
	}

	spec, err := cat.Get("humanizer", "base")
	if err != nil {
		t.Fatalf("Get(humanizer, base) error = %v", err)
	}
	if spec.Kind != "humanizer" || spec.Name != "base" {
		t.Fatalf("unexpected spec: %#v", spec)
	}
}

func TestCatalogRenderReplacesVariables(t *testing.T) {
	ResetDefaultCatalogForTests()
	t.Chdir(t.TempDir())

	cat, err := DefaultCatalog()
	if err != nil {
		t.Fatalf("DefaultCatalog() error = %v", err)
	}

	rendered, spec, err := cat.Render("image", "cover-default", map[string]string{
		"ARTICLE_TITLE":   "测试标题",
		"ARTICLE_SUMMARY": "测试摘要",
		"VISUAL_STYLE":    "极简",
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if spec.Name != "cover-default" {
		t.Fatalf("spec.Name = %q", spec.Name)
	}
	if !strings.Contains(rendered, "测试标题") || !strings.Contains(rendered, "极简") {
		t.Fatalf("rendered prompt = %q", rendered)
	}
}

func TestCatalogPrefersExplicitPromptDirOverBuiltin(t *testing.T) {
	ResetDefaultCatalogForTests()
	tmpDir := t.TempDir()
	overrideDir := filepath.Join(tmpDir, "prompts", "humanizer")
	if err := os.MkdirAll(overrideDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	override := strings.Join([]string{
		"name: medium",
		"kind: humanizer",
		"description: override",
		"version: \"1.0\"",
		"template: |",
		"  override medium",
	}, "\n")
	if err := os.WriteFile(filepath.Join(overrideDir, "medium.yaml"), []byte(override), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	t.Setenv(promptsDirEnvVar, filepath.Join(tmpDir, "prompts"))

	cat := NewCatalog()
	if err := cat.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	spec, err := cat.Get("humanizer", "medium")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if spec.Source != filepath.Join(tmpDir, "prompts") {
		t.Fatalf("Source = %q", spec.Source)
	}
	if strings.TrimSpace(spec.Template) != "override medium" {
		t.Fatalf("Template = %q", spec.Template)
	}
}
