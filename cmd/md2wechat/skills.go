package main

import (
	"fmt"
	"os"

	md2wechatskill "github.com/geekjourneyx/md2wechat-skill"
	"github.com/geekjourneyx/md2wechat-skill/internal/skillcontent"
	"github.com/spf13/cobra"
)

const (
	codeSkillsShown = "SKILLS_SHOWN"
	codeSkillsRead  = "SKILL_READ"
)

var skillsReadJSON bool

type skillsReadData struct {
	Skill    string `json:"skill"`
	Path     string `json:"path"`
	Content  string `json:"content"`
	Guidance string `json:"guidance,omitempty"`
}

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Read embedded agent skill content",
	Long: "Read agent-readable skill content embedded in this CLI binary, " +
		"so agents can discover the current capabilities and SOP without relying on README files or network access.",
}

var skillsListCmd = &cobra.Command{
	Use:   "list [name[/path]]",
	Short: "List embedded skills or files under one skill path",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return newCLIError(codeConfigInvalid, "skills list takes at most 1 argument: [name[/path]]")
		}
		reader, err := newSkillReader()
		if err != nil {
			return wrapCLIError(codeError, err, err.Error())
		}
		if len(args) == 0 {
			skills, err := reader.List()
			if err != nil {
				return wrapCLIError(codeError, err, err.Error())
			}
			responseSuccessWith(codeSkillsShown, "Skills shown", map[string]any{
				"skills": skills,
				"count":  len(skills),
			})
			return nil
		}

		entries, listed, err := reader.ListPath(args[0])
		if err != nil {
			return newCLIError(codeConfigInvalid, err.Error())
		}
		responseSuccessWith(codeSkillsShown, "Skill path shown", map[string]any{
			"path":    listed,
			"entries": entries,
			"count":   len(entries),
		})
		return nil
	},
}

var skillsReadCmd = &cobra.Command{
	Use:   "read <name>[/<path>] [path]",
	Short: "Read a skill's SKILL.md or a file under that skill",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, relpath, err := parseSkillReadTarget(args)
		if err != nil {
			return err
		}
		reader, err := newSkillReader()
		if err != nil {
			return wrapCLIError(codeError, err, err.Error())
		}

		var content []byte
		pathOut := "SKILL.md"
		if relpath == "" {
			content, err = reader.ReadSkill(name)
		} else {
			content, pathOut, err = reader.ReadReference(name, relpath)
		}
		if err != nil {
			return newCLIError(codeConfigInvalid, err.Error())
		}

		if skillsReadJSON || jsonOutput {
			data := skillsReadData{Skill: name, Path: pathOut, Content: string(content)}
			if pathOut == "SKILL.md" {
				data.Guidance = skillReadGuidance(name)
			}
			responseSuccessWith(codeSkillsRead, "Skill read", data)
			return nil
		}

		if _, err := os.Stdout.Write(content); err != nil {
			return wrapCLIError(codeError, err, "failed to write skill content")
		}
		return nil
	},
}

func init() {
	skillsReadCmd.Flags().BoolVar(&skillsReadJSON, "json", false, "output as a JSON envelope instead of raw markdown")
	skillsListCmd.Flags().Bool("json", false, "accepted for symmetry; list output is always JSON")
	skillsCmd.AddCommand(skillsListCmd, skillsReadCmd)
}

func newSkillReader() (*skillcontent.Reader, error) {
	fsys, err := md2wechatskill.EmbeddedSkillContent()
	if err != nil {
		return nil, fmt.Errorf("skill content not embedded in this build: %w", err)
	}
	return skillcontent.New(fsys), nil
}

func parseSkillReadTarget(args []string) (name, relpath string, err error) {
	switch len(args) {
	case 1:
		name, relpath = skillcontent.SplitArg(args[0])
		return name, relpath, nil
	case 2:
		return args[0], args[1], nil
	default:
		return "", "", newCLIError(codeConfigInvalid, "skills read requires 1 or 2 arguments: <name>[/<path>] [path]")
	}
}

func skillReadGuidance(name string) string {
	return fmt.Sprintf("Read this skill's own files with `md2wechat skills read %s <relative-path>` so the SOP stays in sync with this CLI version.", name)
}
