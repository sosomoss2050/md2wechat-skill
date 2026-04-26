package layoutcatalog

import (
	"regexp"
	"strings"
)

type ValidationIssue struct {
	Module  string `json:"module"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
}

type ValidationReport struct {
	Errors   []ValidationIssue `json:"errors"`
	Warnings []ValidationIssue `json:"warnings"`
}

var blockOpenRE = regexp.MustCompile(`^:::([a-z][a-z0-9_-]*)\s*$`)

func (c *Catalog) Validate(markdown string) ValidationReport {
	var r ValidationReport
	lines := strings.Split(markdown, "\n")
	i := 0
	for i < len(lines) {
		m := blockOpenRE.FindStringSubmatch(strings.TrimRight(lines[i], "\r"))
		if m == nil {
			i++
			continue
		}
		moduleName := m[1]
		startLine := i + 1
		j := i + 1
		body := []string{}
		for j < len(lines) && strings.TrimRight(lines[j], "\r") != ":::" {
			body = append(body, lines[j])
			j++
		}
		if j >= len(lines) {
			r.Errors = append(r.Errors, ValidationIssue{
				Module:  moduleName,
				Line:    startLine,
				Message: "unterminated :::" + moduleName + " block",
			})
			break
		}
		c.validateBlock(moduleName, body, startLine, &r)
		i = j + 1
	}
	return r
}

func (c *Catalog) validateBlock(name string, body []string, line int, r *ValidationReport) {
	spec, ok := c.Get(name)
	if !ok {
		r.Warnings = append(r.Warnings, ValidationIssue{
			Module:  name,
			Line:    line,
			Message: "unknown layout module (CLI catalog may be older than the API)",
		})
		return
	}
	present := map[string]string{}
	for _, ln := range body {
		ln = strings.TrimRight(ln, "\r")
		if strings.TrimSpace(ln) == "" {
			continue
		}
		idx := strings.Index(ln, ":")
		if idx <= 0 {
			continue // row line (rows mode) — skip strict parse
		}
		k := strings.TrimSpace(ln[:idx])
		v := strings.TrimSpace(ln[idx+1:])
		present[k] = v
	}
	if spec.Fields != nil {
		for _, f := range spec.Fields.Required {
			if v, ok := present[f.Name]; !ok || v == "" {
				r.Errors = append(r.Errors, ValidationIssue{
					Module:  name,
					Field:   f.Name,
					Line:    line,
					Message: "required field missing",
				})
			}
		}
		for _, f := range append(spec.Fields.Required, spec.Fields.Optional...) {
			if v, ok := present[f.Name]; ok && v != "" {
				if err := checkEnum(f, v); err != nil {
					r.Errors = append(r.Errors, ValidationIssue{
						Module:  name,
						Field:   f.Name,
						Line:    line,
						Message: err.Error(),
					})
				}
			}
		}
	}
}
