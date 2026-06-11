package skillcontent

import (
	"strings"
	"testing"
	"testing/fstest"
)

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"md2wechat/SKILL.md":             {Data: []byte("---\nname: md2wechat\nversion: 1.2.3\ndescription: \"Convert Markdown\"\nmetadata:\n  cliHelp: \"md2wechat --help\"\n---\nbody\n")},
		"md2wechat/references/agent.md":  {Data: []byte("# Agent")},
		"md2wechat/references/style.md":  {Data: []byte("# Style")},
		"md2wechat/assets/example.json":  {Data: []byte("{}")},
		"not-a-skill/references/junk.md": {Data: []byte("junk")},
	}
}

func TestListReturnsSkillFrontmatter(t *testing.T) {
	reader := New(testFS())

	skills, err := reader.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(skills) != 1 {
		t.Fatalf("len(skills) = %d, want 1", len(skills))
	}
	got := skills[0]
	if got.Name != "md2wechat" || got.Description != "Convert Markdown" || got.Version != "1.2.3" {
		t.Fatalf("skill metadata = %#v", got)
	}
	if got.Metadata["cliHelp"] != "md2wechat --help" {
		t.Fatalf("metadata = %#v", got.Metadata)
	}
}

func TestListPathListsOneLayerWithSkillPrefixedPaths(t *testing.T) {
	reader := New(testFS())

	entries, listed, err := reader.ListPath("md2wechat")
	if err != nil {
		t.Fatalf("ListPath() error = %v", err)
	}

	if listed != "md2wechat" {
		t.Fatalf("listed = %q, want md2wechat", listed)
	}
	if len(entries) != 3 {
		t.Fatalf("entries = %#v", entries)
	}
	if entries[0].Path != "md2wechat/SKILL.md" || entries[0].IsDir {
		t.Fatalf("entry[0] = %#v", entries[0])
	}
	if entries[2].Path != "md2wechat/references" || !entries[2].IsDir {
		t.Fatalf("entry[2] = %#v", entries[2])
	}

	subEntries, subListed, err := reader.ListPath("md2wechat/references")
	if err != nil {
		t.Fatalf("ListPath(subdir) error = %v", err)
	}
	if subListed != "md2wechat/references" || len(subEntries) != 2 {
		t.Fatalf("sub listing = %q %#v", subListed, subEntries)
	}
}

func TestReadSkillAndReferenceRejectTraversal(t *testing.T) {
	reader := New(testFS())

	main, err := reader.ReadSkill("md2wechat")
	if err != nil {
		t.Fatalf("ReadSkill() error = %v", err)
	}
	if !strings.HasPrefix(string(main), "---\nname: md2wechat") {
		t.Fatalf("main content = %q", string(main))
	}

	ref, cleaned, err := reader.ReadReference("md2wechat", "references/agent.md")
	if err != nil {
		t.Fatalf("ReadReference() error = %v", err)
	}
	if string(ref) != "# Agent" || cleaned != "references/agent.md" {
		t.Fatalf("reference = %q cleaned=%q", string(ref), cleaned)
	}

	for _, bad := range []string{"../../etc/passwd", "/etc/passwd", "..", "", "references/../../x", `..\x`} {
		data, _, err := reader.ReadReference("md2wechat", bad)
		if err == nil {
			t.Fatalf("ReadReference(%q) expected error", bad)
		}
		if data != nil {
			t.Fatalf("ReadReference(%q) leaked data %q", bad, string(data))
		}
	}
}

func TestUnknownSkillAndInvalidListPathReturnErrors(t *testing.T) {
	reader := New(testFS())

	if _, err := reader.ReadSkill("nope"); err == nil || !strings.Contains(err.Error(), "unknown skill") {
		t.Fatalf("ReadSkill unknown error = %v", err)
	}
	if _, _, err := reader.ListPath("md2wechat/../../etc"); err == nil || !strings.Contains(err.Error(), "invalid path") {
		t.Fatalf("ListPath traversal error = %v", err)
	}
	if _, _, err := reader.ListPath("md2wechat/SKILL.md"); err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("ListPath file error = %v", err)
	}
}
