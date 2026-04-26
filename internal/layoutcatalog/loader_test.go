package layoutcatalog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadBuiltinIncludesHero(t *testing.T) {
	c := NewCatalog()
	if err := c.Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	spec, ok := c.Get("hero")
	if !ok {
		t.Fatalf("expected hero module to be present")
	}
	if spec.Category != "opening" {
		t.Errorf("hero.Category = %q, want opening", spec.Category)
	}
}

func TestEnvOverrideTrumpsBuiltin(t *testing.T) {
	ResetDefaultCatalogForTests()
	t.Cleanup(ResetDefaultCatalogForTests)

	dir := t.TempDir()
	override := filepath.Join(dir, "opening")
	if err := os.MkdirAll(override, 0o755); err != nil {
		t.Fatal(err)
	}
	yaml := []byte(`schema_version: "1"
name: hero
version: "999.0.0"
category: opening
serves: [attention]
metadata:
  author: test
  provenance: override
`)
	if err := os.WriteFile(filepath.Join(override, "hero.yaml"), yaml, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MD2WECHAT_LAYOUT_DIR", dir)

	c := NewCatalog()
	if err := c.Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	spec, _ := c.Get("hero")
	if spec.Version != "999.0.0" {
		t.Errorf("override not applied: version = %q", spec.Version)
	}
}

func TestParseLayoutSpecRejectsInvalidServes(t *testing.T) {
	yaml := []byte(`schema_version: "1"
name: bad
version: "1.0.0"
category: opening
serves: [bogus]
metadata:
  author: x
  provenance: builtin
`)
	_, err := parseLayoutSpec(yaml)
	if err == nil {
		t.Fatal("expected error for invalid serves value")
	}
}

func TestEnvOverrideBeatsLocalDir(t *testing.T) {
	localDir := t.TempDir()
	localOpening := filepath.Join(localDir, "opening")
	if err := os.MkdirAll(localOpening, 0o755); err != nil {
		t.Fatal(err)
	}
	localYAML := []byte(`schema_version: "1"
name: hero
version: "2.0.0"
category: opening
serves: [attention]
metadata:
  author: local
  provenance: local
`)
	if err := os.WriteFile(filepath.Join(localOpening, "hero.yaml"), localYAML, 0o644); err != nil {
		t.Fatal(err)
	}

	envDir := t.TempDir()
	envOpening := filepath.Join(envDir, "opening")
	if err := os.MkdirAll(envOpening, 0o755); err != nil {
		t.Fatal(err)
	}
	envYAML := []byte(`schema_version: "1"
name: hero
version: "3.0.0"
category: opening
serves: [attention]
metadata:
  author: env
  provenance: env
`)
	if err := os.WriteFile(filepath.Join(envOpening, "hero.yaml"), envYAML, 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("MD2WECHAT_LAYOUT_DIR", envDir)
	ResetDefaultCatalogForTests()
	t.Cleanup(ResetDefaultCatalogForTests)

	c := NewCatalog()
	if err := c.Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	spec, _ := c.Get("hero")
	if spec.Version != "3.0.0" {
		t.Errorf("env override should win, got version %q", spec.Version)
	}
}

func TestAllBuiltinModulesLoadCleanly(t *testing.T) {
	c := NewCatalog()
	if err := c.Load(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if got := len(c.modules); got < 38 {
		t.Errorf("expected at least 38 modules, got %d", got)
	}
	for name, m := range c.modules {
		if m.Metadata.Provenance == "" {
			t.Errorf("%s missing provenance", name)
		}
		if m.Metadata.InspiredBy == "" && m.Metadata.Provenance == "builtin" {
			t.Errorf("%s builtin module missing inspired_by", name)
		}
	}
}
