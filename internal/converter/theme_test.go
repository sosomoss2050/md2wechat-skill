package converter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestThemeManagerLoadsBuiltinThemesWithoutFilesystemAssets(t *testing.T) {
	t.Chdir(t.TempDir())

	tm := NewThemeManager()
	theme, err := tm.GetTheme("default")
	if err != nil {
		t.Fatalf("GetTheme(default) error = %v", err)
	}
	if theme.Name != "default" || theme.Type != "api" {
		t.Fatalf("unexpected builtin theme: %#v", theme)
	}

	aiTheme, err := tm.GetTheme("autumn-warm")
	if err != nil {
		t.Fatalf("GetTheme(autumn-warm) error = %v", err)
	}
	if aiTheme.Type != "ai" {
		t.Fatalf("expected AI builtin theme, got %#v", aiTheme)
	}
}

func TestThemeManagerParsesStyleMetadata(t *testing.T) {
	t.Chdir(t.TempDir())

	tm := NewThemeManager()
	theme, err := tm.GetTheme("minimal-blue")
	if err != nil {
		t.Fatalf("GetTheme(minimal-blue) error = %v", err)
	}
	if theme.Style.Series != "minimal" {
		t.Fatalf("Style.Series = %q, want minimal", theme.Style.Series)
	}
	if theme.Style.Color != "blue" {
		t.Fatalf("Style.Color = %q, want blue", theme.Style.Color)
	}
	if theme.Style.Mood == "" {
		t.Fatal("expected Style.Mood")
	}
	if theme.Style.BestFor == "" {
		t.Fatal("expected Style.BestFor")
	}
}

func TestThemeManagerExpandsAPICollectionThemes(t *testing.T) {
	t.Chdir(t.TempDir())

	tm := NewThemeManager()
	theme, err := tm.GetTheme("elegant-green")
	if err != nil {
		t.Fatalf("GetTheme(elegant-green) error = %v", err)
	}
	if theme.Type != "api" || theme.APITheme != "elegant-green" || !theme.Selectable() {
		t.Fatalf("unexpected expanded API theme: %#v", theme)
	}
	if theme.Style.Series != "elegant" || theme.Style.Color != "green" || theme.Style.Layout == "" {
		t.Fatalf("unexpected expanded API theme style: %#v", theme.Style)
	}
}

func TestThemeManagerExpandsAPICollectionFeaturedThemes(t *testing.T) {
	t.Chdir(t.TempDir())

	tm := NewThemeManager()
	tests := []struct {
		name  string
		color string
	}{
		{name: "sspai-red", color: "red"},
		{name: "wechat-native", color: "green"},
	}

	for _, tt := range tests {
		theme, err := tm.GetTheme(tt.name)
		if err != nil {
			t.Fatalf("GetTheme(%s) error = %v", tt.name, err)
		}
		if theme.Type != "api" || theme.APITheme != tt.name || !theme.Selectable() {
			t.Fatalf("unexpected expanded featured theme: %#v", theme)
		}
		if theme.Style.Series != "featured" || theme.Style.Color != tt.color {
			t.Fatalf("unexpected expanded featured theme style: %#v", theme.Style)
		}
	}
}

func TestThemeSelectabilityMarksAPICollectionFalse(t *testing.T) {
	t.Chdir(t.TempDir())

	tm := NewThemeManager()
	theme, err := tm.GetTheme("api-collection")
	if err != nil {
		t.Fatalf("GetTheme(api-collection) error = %v", err)
	}
	if theme.Selectable() {
		t.Fatal("expected api-collection not to be selectable")
	}
}

func TestThemeMetadataIncompleteRequiresAllCoreStyleFields(t *testing.T) {
	partial := Theme{
		Description: "partially documented theme",
		Style: ThemeStyle{
			Series: "minimal",
		},
	}
	if !partial.MetadataIncomplete() {
		t.Fatal("expected partially populated style metadata to be incomplete")
	}

	complete := Theme{
		Description: "fully documented theme",
		Style: ThemeStyle{
			Series:  "minimal",
			Color:   "blue",
			Mood:    "clean",
			BestFor: "technical docs",
		},
	}
	if complete.MetadataIncomplete() {
		t.Fatal("expected fully populated style metadata to be complete")
	}
}

func TestThemeManagerPrefersCurrentDirectoryThemeOverBuiltin(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	if err := os.MkdirAll("themes", 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	override := []byte("name: default\ntype: api\ndescription: local override\napi_theme: default\n")
	if err := os.WriteFile(filepath.Join("themes", "default.yaml"), override, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	tm := NewThemeManager()
	theme, err := tm.GetTheme("default")
	if err != nil {
		t.Fatalf("GetTheme(default) error = %v", err)
	}
	if theme.Description != "local override" {
		t.Fatalf("Description = %q, want local override", theme.Description)
	}
}

func TestThemeManagerPrefersExplicitThemesDirOverBuiltin(t *testing.T) {
	customDir := filepath.Join(t.TempDir(), "custom-themes")
	if err := os.MkdirAll(customDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	override := []byte("name: default\ntype: api\ndescription: env override\napi_theme: default\n")
	if err := os.WriteFile(filepath.Join(customDir, "default.yaml"), override, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	t.Setenv(themesDirEnvVar, customDir)
	t.Chdir(t.TempDir())

	tm := NewThemeManager()
	theme, err := tm.GetTheme("default")
	if err != nil {
		t.Fatalf("GetTheme(default) error = %v", err)
	}
	if theme.Description != "env override" {
		t.Fatalf("Description = %q, want env override", theme.Description)
	}
}
