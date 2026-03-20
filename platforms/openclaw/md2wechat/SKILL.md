---
name: md2wechat
description: Convert Markdown to WeChat Official Account HTML, process images, generate cover and infographic images, and create drafts for OpenClaw.
metadata: {"openclaw":{"emoji":"📝","homepage":"https://github.com/geekjourneyx/md2wechat-skill","primaryEnv":"WECHAT_APPID","requires":{"env":["WECHAT_APPID","WECHAT_SECRET"]},"install":[{"id":"openclaw-installer-shell","kind":"download","label":"Download fixed-version OpenClaw installer (shell)","url":"https://github.com/geekjourneyx/md2wechat-skill/releases/download/v2.0.0/install-openclaw.sh","os":["darwin","linux"]},{"id":"openclaw-installer-powershell","kind":"download","label":"Download fixed-version installer (PowerShell)","url":"https://github.com/geekjourneyx/md2wechat-skill/releases/download/v2.0.0/install.ps1","os":["win32"]},{"id":"openclaw-skill-bundle","kind":"download","label":"Download OpenClaw skill bundle","url":"https://github.com/geekjourneyx/md2wechat-skill/releases/download/v2.0.0/md2wechat-openclaw-skill.tar.gz","archive":"tar.gz","targetDir":"~/.openclaw/skills","os":["darwin","linux","win32"]},{"id":"openclaw-runtime-linux","kind":"download","label":"Download md2wechat runtime (Linux amd64)","url":"https://github.com/geekjourneyx/md2wechat-skill/releases/download/v2.0.0/md2wechat-linux-amd64","targetDir":"~/.openclaw/tools/md2wechat","os":["linux"]},{"id":"openclaw-runtime-darwin","kind":"download","label":"Download md2wechat runtime (macOS amd64)","url":"https://github.com/geekjourneyx/md2wechat-skill/releases/download/v2.0.0/md2wechat-darwin-amd64","targetDir":"~/.openclaw/tools/md2wechat","os":["darwin"]}]}}
---

# md2wechat for OpenClaw

透明披露：

- 会读取本地 Markdown 文件和本地图片。
- 可能把处理后的图片和 HTML 上传到微信草稿箱或素材接口。
- 可能调用外部图像服务来生成图片或补图。
- 草稿上传需要 `WECHAT_APPID` 和 `WECHAT_SECRET`。
- 图片生成通常还需要 `IMAGE_API_KEY`，以及可选的 `IMAGE_PROVIDER` / `IMAGE_API_BASE`。

配置入口：

- 默认先检查 `~/.config/md2wechat/config.yaml`
- 如需切换 API 域名、图片 provider 或确认默认模式，先看仓库文档 `docs/CONFIG.md`
- 未显式传 `--mode` 时，`convert` 默认仍走 `api`

## 运行边界

- 本 skill 只执行已经安装好的 `md2wechat` runtime。
- 优先查找 `~/.openclaw/tools/md2wechat` 下的 runtime，然后查找 `PATH`。
- 运行时不会下载二进制，不会写入缓存下载器，也不会静默回退到远程拉取。
- `metadata.openclaw.install` 当前提供的是固定版本安装资源与 installer 入口，真正完整的自动安装主线仍以 `install-openclaw.sh` 为准。

## 推荐流程

1. 先通过固定版本 OpenClaw installer 完成 skill 和 runtime 安装。
2. 先用发现命令确认当前实例支持的能力和资源：
   - `md2wechat capabilities --json`
   - `md2wechat providers list --json`
   - `md2wechat themes list --json`
   - `md2wechat prompts list --json`
   - `md2wechat prompts list --kind image --archetype cover --json`
3. 再执行以下任务：
   - `convert <file.md> --preview`
   - `convert <file.md> --draft --cover <cover.jpg>`
   - `generate_cover --article <file.md>`
   - `generate_infographic --article <file.md> --preset infographic-comparison`
   - `generate_image --preset cover-hero --article <file.md> --model <image-model>`
   - `create_image_post -m <file.md> -t "<title>"`
4. 如果要使用 AI 转换或 AI 图片，再补齐图像服务配置。

当任务依赖具体资源时，先检查再执行：

- `md2wechat providers show <name> --json`
- `md2wechat themes show <name> --json`
- `md2wechat prompts show <name> --kind <kind> --json`
- `md2wechat prompts render <name> --kind <kind> --var KEY=VALUE --json`

高频图片工作流建议：

- 封面图优先用 `generate_cover`
- 信息图优先用 `generate_infographic`
- 只有在没有合适 preset 时，再直接用 `generate_image "raw prompt"`

See [references/runtime.md](references/runtime.md) for the runtime lookup contract.
