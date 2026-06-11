package skillcontent

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Reader struct {
	fsys fs.FS
}

func New(fsys fs.FS) *Reader {
	return &Reader{fsys: fsys}
}

type SkillInfo struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Version     string         `json:"version,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type DirEntry struct {
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
}

func (r *Reader) List() ([]SkillInfo, error) {
	if r == nil || r.fsys == nil {
		return nil, fmt.Errorf("skill content not embedded in this build")
	}
	entries, err := fs.ReadDir(r.fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read embedded skills: %w", err)
	}
	skills := make([]SkillInfo, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, ok := r.skillInfo(entry.Name())
		if ok {
			skills = append(skills, info)
		}
	}
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Name < skills[j].Name
	})
	return skills, nil
}

func (r *Reader) ListPath(arg string) ([]DirEntry, string, error) {
	name, subpath := SplitArg(arg)
	if err := r.ensureSkill(name); err != nil {
		return nil, "", err
	}

	dir := name
	if subpath != "" {
		cleaned, err := cleanSubPath(subpath)
		if err != nil {
			return nil, "", err
		}
		dir = name + "/" + cleaned
		info, err := fs.Stat(r.fsys, dir)
		if err != nil {
			return nil, "", fmt.Errorf("path %q not found in skill %q", subpath, name)
		}
		if !info.IsDir() {
			return nil, "", fmt.Errorf("path %q is not a directory; use skills read", subpath)
		}
	}

	entries, err := fs.ReadDir(r.fsys, dir)
	if err != nil {
		return nil, "", fmt.Errorf("read embedded skill path: %w", err)
	}
	out := make([]DirEntry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, DirEntry{Path: dir + "/" + entry.Name(), IsDir: entry.IsDir()})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Path < out[j].Path
	})
	return out, dir, nil
}

func SplitArg(arg string) (name, rest string) {
	name, rest, _ = strings.Cut(arg, "/")
	return name, rest
}

func (r *Reader) ReadSkill(name string) ([]byte, error) {
	if err := r.ensureSkill(name); err != nil {
		return nil, err
	}
	data, err := fs.ReadFile(r.fsys, name+"/SKILL.md")
	if err != nil {
		return nil, fmt.Errorf("read embedded skill %q: %w", name, err)
	}
	return data, nil
}

func (r *Reader) ReadReference(name, relpath string) ([]byte, string, error) {
	if err := r.ensureSkill(name); err != nil {
		return nil, "", err
	}
	cleaned, err := cleanSubPath(relpath)
	if err != nil {
		return nil, "", err
	}
	fullPath := name + "/" + cleaned
	info, err := fs.Stat(r.fsys, fullPath)
	if err != nil {
		return nil, "", fmt.Errorf("reference %q not found in skill %q", relpath, name)
	}
	if info.IsDir() {
		return nil, "", fmt.Errorf("reference %q is a directory, not a file", relpath)
	}
	data, err := fs.ReadFile(r.fsys, fullPath)
	if err != nil {
		return nil, "", fmt.Errorf("read embedded skill reference %q: %w", relpath, err)
	}
	return data, cleaned, nil
}

func (r *Reader) ensureSkill(name string) error {
	if r == nil || r.fsys == nil {
		return fmt.Errorf("skill content not embedded in this build")
	}
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return unknownSkill(name)
	}
	info, err := fs.Stat(r.fsys, name)
	if err != nil || !info.IsDir() {
		return unknownSkill(name)
	}
	if _, err := fs.Stat(r.fsys, name+"/SKILL.md"); err != nil {
		return unknownSkill(name)
	}
	return nil
}

func (r *Reader) skillInfo(name string) (SkillInfo, bool) {
	data, err := fs.ReadFile(r.fsys, name+"/SKILL.md")
	if err != nil {
		return SkillInfo{}, false
	}
	description, version, metadata := parseFrontmatter(data)
	return SkillInfo{
		Name:        name,
		Description: description,
		Version:     version,
		Metadata:    metadata,
	}, true
}

func unknownSkill(name string) error {
	return fmt.Errorf("unknown skill %q; run 'md2wechat skills list' to see available skills", name)
}

func cleanSubPath(relpath string) (string, error) {
	cleaned := path.Clean(relpath)
	if relpath == "" || path.IsAbs(relpath) || cleaned == "." || cleaned == ".." ||
		strings.HasPrefix(cleaned, "../") || strings.HasPrefix(cleaned, `..\`) ||
		strings.Contains(cleaned, `/../`) || strings.Contains(cleaned, `\..\`) {
		return "", fmt.Errorf("invalid path %q: must be a relative path without '..'", relpath)
	}
	return cleaned, nil
}

func parseFrontmatter(skillMD []byte) (description, version string, metadata map[string]any) {
	lines := strings.Split(string(skillMD), "\n")
	if len(lines) == 0 || strings.TrimRight(lines[0], "\r") != "---" {
		return "", "", nil
	}

	block := make([]string, 0, len(lines))
	closed := false
	for _, line := range lines[1:] {
		if strings.TrimRight(line, "\r") == "---" {
			closed = true
			break
		}
		block = append(block, line)
	}
	if !closed {
		return "", "", nil
	}

	var frontmatter struct {
		Description string         `yaml:"description"`
		Version     string         `yaml:"version"`
		Metadata    map[string]any `yaml:"metadata"`
	}
	if err := yaml.Unmarshal([]byte(strings.Join(block, "\n")), &frontmatter); err != nil {
		return "", "", nil
	}
	return frontmatter.Description, frontmatter.Version, frontmatter.Metadata
}
