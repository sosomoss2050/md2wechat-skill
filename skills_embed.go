package md2wechatskill

import (
	"embed"
	"io/fs"
)

//go:embed skills/*
var embeddedSkills embed.FS

func EmbeddedSkillContent() (fs.FS, error) {
	return fs.Sub(embeddedSkills, "skills")
}
