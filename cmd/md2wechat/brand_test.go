package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/geekjourneyx/md2wechat-skill/internal/action"
)

// parseBrandJSON 解析 brand 命令的 JSON 输出
func parseBrandJSON(t *testing.T, output []byte) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("invalid JSON output: %v\nOutput: %s", err, output)
	}
	return result
}

// ============ init group (4 tests) ============

// TestBrandInit_CreatesFile init on empty dir creates brand.md, returns BRAND_INITIALIZED
func TestBrandInit_CreatesFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	stdout := captureStdout(t, func() {
		if err := runBrandInit(); err != nil {
			t.Fatalf("runBrandInit() error = %v", err)
		}
	})

	result := parseBrandJSON(t, stdout)

	// 检查 JSON envelope
	if result["success"] != true {
		t.Errorf("expected success=true, got %v", result["success"])
	}
	if result["code"] != "BRAND_INITIALIZED" {
		t.Errorf("expected code=BRAND_INITIALIZED, got %v", result["code"])
	}
	if result["schema_version"] != action.SchemaVersion {
		t.Errorf("expected schema_version=%s, got %v", action.SchemaVersion, result["schema_version"])
	}
	if result["status"] != string(action.StatusCompleted) {
		t.Errorf("expected status=%s, got %v", action.StatusCompleted, result["status"])
	}

	// 检查文件是否创建
	brandPath := filepath.Join(tmpHome, ".config", "md2wechat", "brand.md")
	if _, err := os.Stat(brandPath); os.IsNotExist(err) {
		t.Fatalf("brand.md not created at %s", brandPath)
	}

	// 检查文件内容不为空
	content, err := os.ReadFile(brandPath)
	if err != nil {
		t.Fatalf("failed to read created brand.md: %v", err)
	}
	if len(content) == 0 {
		t.Error("brand.md is empty")
	}
}

// TestBrandInit_Idempotent init twice, second call still returns BRAND_INITIALIZED (not error), file not overwritten
func TestBrandInit_Idempotent(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// 第一次 init
	stdout1 := captureStdout(t, func() {
		if err := runBrandInit(); err != nil {
			t.Fatalf("first runBrandInit() error = %v", err)
		}
	})

	result1 := parseBrandJSON(t, stdout1)
	if result1["code"] != "BRAND_INITIALIZED" {
		t.Errorf("first init: expected code=BRAND_INITIALIZED, got %v", result1["code"])
	}

	brandPath := filepath.Join(tmpHome, ".config", "md2wechat", "brand.md")

	// 修改文件内容，标记一下
	testContent := "# MODIFIED BY TEST\n\nCustom content here.\n"
	if err := os.WriteFile(brandPath, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to modify brand.md: %v", err)
	}

	// 第二次 init
	stdout2 := captureStdout(t, func() {
		if err := runBrandInit(); err != nil {
			t.Fatalf("second runBrandInit() error = %v", err)
		}
	})

	result2 := parseBrandJSON(t, stdout2)
	if result2["success"] != true {
		t.Errorf("second init: expected success=true, got %v", result2["success"])
	}
	if result2["code"] != "BRAND_INITIALIZED" {
		t.Errorf("second init: expected code=BRAND_INITIALIZED, got %v", result2["code"])
	}

	// 检查文件没有被覆盖
	content, err := os.ReadFile(brandPath)
	if err != nil {
		t.Fatalf("failed to read brand.md after second init: %v", err)
	}
	if string(content) != testContent {
		t.Error("brand.md was overwritten by second init (should be idempotent)")
	}
}

// TestBrandInit_JSONEnvelope output is valid JSON with schema_version:"v1", success:true, status:"completed"
func TestBrandInit_JSONEnvelope(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	stdout := captureStdout(t, func() {
		if err := runBrandInit(); err != nil {
			t.Fatalf("runBrandInit() error = %v", err)
		}
	})

	result := parseBrandJSON(t, stdout)

	// 检查 JSON envelope 契约
	if result["schema_version"] != "v1" {
		t.Errorf("expected schema_version=v1, got %v", result["schema_version"])
	}
	if result["success"] != true {
		t.Errorf("expected success=true, got %v", result["success"])
	}
	if result["status"] != "completed" {
		t.Errorf("expected status=completed, got %v", result["status"])
	}

	// 检查必须的字段存在
	if _, ok := result["code"]; !ok {
		t.Error("missing 'code' field in JSON envelope")
	}
	if _, ok := result["message"]; !ok {
		t.Error("missing 'message' field in JSON envelope")
	}
}

// TestBrandInit_CreatesParentDir init when ~/.config/md2wechat/ doesn't exist, creates dir + file
func TestBrandInit_CreatesParentDir(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// 确保父目录不存在
	configDir := filepath.Join(tmpHome, ".config", "md2wechat")
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Fatalf("config dir already exists (test precondition failed)")
	}

	stdout := captureStdout(t, func() {
		if err := runBrandInit(); err != nil {
			t.Fatalf("runBrandInit() error = %v", err)
		}
	})

	result := parseBrandJSON(t, stdout)
	if result["code"] != "BRAND_INITIALIZED" {
		t.Errorf("expected code=BRAND_INITIALIZED, got %v", result["code"])
	}

	// 检查父目录和文件都被创建
	if info, err := os.Stat(configDir); err != nil {
		t.Fatalf("config dir was not created: %v", err)
	} else if !info.IsDir() {
		t.Fatal("config path exists but is not a directory")
	}

	brandPath := filepath.Join(configDir, "brand.md")
	if _, err := os.Stat(brandPath); os.IsNotExist(err) {
		t.Fatalf("brand.md was not created at %s", brandPath)
	}
}

// ============ show group (4 tests) ============

// TestBrandShow_NotFound show when no file → BRAND_NOT_FOUND, success:false, status:"action_required"
func TestBrandShow_NotFound(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// 确保文件不存在
	brandPath := filepath.Join(tmpHome, ".config", "md2wechat", "brand.md")
	if _, err := os.Stat(brandPath); !os.IsNotExist(err) {
		t.Fatalf("brand.md exists (test precondition failed)")
	}

	stdout := captureStdout(t, func() {
		if err := runBrandShow(); err != nil {
			t.Fatalf("runBrandShow() error = %v", err)
		}
	})

	result := parseBrandJSON(t, stdout)

	// 检查返回值
	if result["success"] != false {
		t.Errorf("expected success=false, got %v", result["success"])
	}
	if result["code"] != "BRAND_NOT_FOUND" {
		t.Errorf("expected code=BRAND_NOT_FOUND, got %v", result["code"])
	}
	if result["status"] != "action_required" {
		t.Errorf("expected status=action_required, got %v", result["status"])
	}
}

// TestBrandShow_ValidFile show after init → BRAND_SHOWN, success:true, data.content present (raw string), data.path uses ~/
func TestBrandShow_ValidFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// 先 init
	captureStdout(t, func() {
		if err := runBrandInit(); err != nil {
			t.Fatalf("runBrandInit() error = %v", err)
		}
	})

	// 再 show
	stdout := captureStdout(t, func() {
		if err := runBrandShow(); err != nil {
			t.Fatalf("runBrandShow() error = %v", err)
		}
	})

	result := parseBrandJSON(t, stdout)

	// 检查返回值
	if result["success"] != true {
		t.Errorf("expected success=true, got %v", result["success"])
	}
	if result["code"] != "BRAND_SHOWN" {
		t.Errorf("expected code=BRAND_SHOWN, got %v", result["code"])
	}

	// 检查 data 字段
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data field is not a map: %T", result["data"])
	}

	// 检查 content 存在且为非空字符串
	content, ok := data["content"].(string)
	if !ok {
		t.Fatalf("data.content is not a string: %T", data["content"])
	}
	if len(content) == 0 {
		t.Error("data.content is empty")
	}

	// 检查 path 使用 ~/ 格式
	path, ok := data["path"].(string)
	if !ok {
		t.Fatalf("data.path is not a string: %T", data["path"])
	}
	if len(path) == 0 {
		t.Error("data.path is empty")
	}
	// path 应该以 ~/ 开头（normalizeBrandPath 的效果）
	if path[0] != '~' && path[0] != '/' {
		t.Errorf("data.path should be normalized to use ~/ or absolute path, got: %s", path)
	}
}

// TestBrandShow_UnreadableFile show when file exists but is unreadable (chmod 000) → BRAND_READ_FAILED
func TestBrandShow_UnreadableFile(t *testing.T) {
	// Skip this test if running as root (root can read 0000 files)
	if os.Getuid() == 0 {
		t.Skip("skipping unreadable file test when running as root")
	}

	oldExit := exitFunc
	t.Cleanup(func() {
		exitFunc = oldExit
	})
	exitFunc = func(code int) {} // no-op exit for test

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// 创建文件，然后设置为不可读
	brandFile := filepath.Join(tmpHome, ".config", "md2wechat", "brand.md")
	if err := os.MkdirAll(filepath.Dir(brandFile), 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	if err := os.WriteFile(brandFile, []byte("# test"), 0644); err != nil {
		t.Fatalf("failed to write brand.md: %v", err)
	}
	if err := os.Chmod(brandFile, 0000); err != nil {
		t.Fatalf("failed to chmod brand.md: %v", err)
	}
	defer func() { _ = os.Chmod(brandFile, 0644) }() // restore for cleanup

	stdout := captureStdout(t, func() {
		if err := runBrandShow(); err != nil {
			responseError(err) // simulate cobra error handling
		}
	})

	result := parseBrandJSON(t, stdout)

	// 检查返回值
	if result["success"] != false {
		t.Errorf("expected success=false, got %v", result["success"])
	}
	if result["code"] != "BRAND_READ_FAILED" {
		t.Errorf("expected code=BRAND_READ_FAILED, got %v", result["code"])
	}
	if result["status"] != "failed" {
		t.Errorf("expected status=failed, got %v", result["status"])
	}
}

// TestBrandShow_JSONEnvelope valid JSON, schema_version:"v1", code:"BRAND_SHOWN", data.path non-empty
func TestBrandShow_JSONEnvelope(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// 先 init
	captureStdout(t, func() {
		if err := runBrandInit(); err != nil {
			t.Fatalf("runBrandInit() error = %v", err)
		}
	})

	// 再 show
	stdout := captureStdout(t, func() {
		if err := runBrandShow(); err != nil {
			t.Fatalf("runBrandShow() error = %v", err)
		}
	})

	result := parseBrandJSON(t, stdout)

	// 检查 JSON envelope 契约
	if result["schema_version"] != "v1" {
		t.Errorf("expected schema_version=v1, got %v", result["schema_version"])
	}
	if result["code"] != "BRAND_SHOWN" {
		t.Errorf("expected code=BRAND_SHOWN, got %v", result["code"])
	}

	// 检查 data.path 非空
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data field is not a map: %T", result["data"])
	}

	path, ok := data["path"].(string)
	if !ok {
		t.Fatalf("data.path is not a string: %T", data["path"])
	}
	if len(path) == 0 {
		t.Error("data.path is empty")
	}
}
