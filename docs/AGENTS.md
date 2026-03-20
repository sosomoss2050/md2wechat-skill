# Agent 协作规范

This repository is a Go CLI project in stabilization. Treat documentation and execution rules as part of the product.

## Working Rules

1. Start by checking the current tree and the exact files that are in scope.
2. Do not revert user changes unless explicitly asked.
3. Prefer `apply_patch` for edits and avoid destructive commands.
4. Keep changes scoped; if a task crosses file boundaries, state the boundary before editing.
5. Preserve CLI and documentation compatibility unless the task explicitly calls for a breaking change.
6. When a task depends on providers, themes, or prompt templates, prefer CLI discovery output over guessing or stale docs.

## Discovery-First Workflow

Before assuming a feature, resource, or prompt exists, inspect the running CLI:

1. `md2wechat capabilities --json`
2. `md2wechat providers list --json`
3. `md2wechat themes list --json`
4. `md2wechat prompts list --json`

Use `show` or `render` only when a task depends on a specific resource:

1. `md2wechat providers show <name> --json`
2. `md2wechat themes show <name> --json`
3. `md2wechat prompts show <name> --kind <kind> --json`
4. `md2wechat prompts render <name> --kind <kind> --var KEY=VALUE --json`

Treat these commands as the source of truth for:

- currently supported image providers
- currently visible themes after override resolution
- bundled and overridden prompt catalog entries

## Verification Order

1. Run `gofmt -l .` when Go files change.
2. Run `go vet ./...` after structural changes.
3. Run `GOCACHE=/tmp/md2wechat-go-build go test ./...` for regression coverage.
4. Run `GOCACHE=/tmp/md2wechat-go-build go test -cover ./...` when coverage is part of the task.
5. If `make release-check` exists in this branch, run it before declaring release-related work done.
6. When release assets or installer scripts change, run artifact smoke and installer smoke against the same bundle before calling the work done.
7. If the task touches release or installer paths, keep the documented primary path versioned and non-`latest`.

## Release And Version Discipline

1. Keep the version source singular and explicit when the repository has one.
2. Keep install scripts, release notes, changelog, and release assets aligned.
3. If the repository does not yet provide a workflow or release gate, document the gap instead of inventing it.
4. Release work is not complete until these are explicitly re-audited and aligned:
   - `.claude-plugin/marketplace.json`
   - `skills/md2wechat/SKILL.md`
   - `platforms/openclaw/md2wechat/SKILL.md`
   - `skills/md2wechat/scripts/run.sh`
   - `scripts/install.sh`
   - `scripts/install-openclaw.sh`
5. If any of the files above still carry an old version, stale release URL, stale maintainer identity, or outdated command examples, block the release and fix them before tagging.

## Documentation Discipline

1. Document only supported paths as the primary path.
2. Label aspirational or not-yet-shipped paths clearly.
3. If a feature is not backed by current code or release assets, do not describe it as shipped.
4. Prefer short, operational instructions over marketing copy.
5. If any CLI command, subcommand, flag, JSON response shape, prompt asset, provider, or theme behavior changes, treat documentation sync as part of the same task.
6. The minimum documentation review set for CLI changes is:
   - `README.md`
   - `docs/DISCOVERY.md`
   - `docs/FAQ.md`
   - the relevant `SKILL.md`
7. If the CLI change affects setup, config, install, or migration behavior, also review:
   - `docs/CONFIG.md`
   - `docs/QUICKSTART.md`
   - `docs/USAGE.md`
   - `docs/OPENCLAW.md` when platform behavior changes
8. Do not stop at “updated one doc”. Re-audit the highest-signal entry points so user-facing docs, agent-facing docs, and CLI behavior stay aligned.
9. When image prompts change, treat both skill entry points as mandatory review targets:
   - `skills/md2wechat/SKILL.md`
   - `platforms/openclaw/md2wechat/SKILL.md`

## Image Prompt Discipline

When adding or changing `internal/assets/builtin/prompts/image/*.yaml`, treat the change as a product contract update, not just a content tweak.

Minimum required fields for builtin image prompts:

- `name`
- `kind: image`
- `description`
- `version`
- `archetype`
- `primary_use_case`
- `recommended_aspect_ratios`
- `default_aspect_ratio`
- `metadata.author`
- `metadata.provenance`
- `template`

Use these when applicable:

- `compatible_use_cases`
- `tags`
- `examples`
- `metadata.inspired_by`

Required constraints:

1. `default_aspect_ratio` must appear in `recommended_aspect_ratios`.
2. If a preset can serve both cover and infographic scenarios, express that through `compatible_use_cases` rather than duplicating the preset.
3. Do not move long image prompts back into Go code when the prompt catalog can carry them.

Required verification after image prompt changes:

1. `gofmt -l .`
2. `GOCACHE=/tmp/md2wechat-go-build go test ./internal/promptcatalog ./cmd/md2wechat`
3. `GOCACHE=/tmp/md2wechat-go-build go test ./...`
4. Re-audit:
   - `README.md`
   - `docs/DISCOVERY.md`
   - `docs/FAQ.md`
   - `skills/md2wechat/SKILL.md`
   - `platforms/openclaw/md2wechat/SKILL.md`

The repository should fail closed on prompt drift: missing primary use case, default aspect ratio, or attribution metadata should be blocked by tests rather than caught manually later.

## Escalation

1. Ask for approval before commands that require network access or writing outside the workspace.
2. If a command fails because of sandboxing, rerun it with the required escalation instead of working around it.
