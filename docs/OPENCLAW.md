# OpenClaw 安装指南

> 本文档对应 OpenClaw 专用 skill 包 `platforms/openclaw/md2wechat/`。
>
> `skills/md2wechat/` 是给 Claude Code / Codex / OpenCode 的 coding-agent skill，两条路径独立维护。
>
> OpenClaw 安装主线是 skill 包与 runtime 一起安装，`run.sh` 只负责启动已安装 runtime，不承担首跑动态下载。

---

## 目录

- [什么是 OpenClaw](#什么是-openclaw)
- [安装方式](#安装方式)
  - [方式一：ClawHub 安装（推荐）](#方式一clawhub-安装推荐)
  - [方式二：一键脚本安装](#方式二一键脚本安装)
  - [方式三：手动安装](#方式三手动安装)
- [配置说明](#配置说明)
- [验证安装](#验证安装)
- [常见问题](#常见问题)
- [与 Claude Code 的区别](#与-claude-code-的区别)

---

## 什么是 OpenClaw

[OpenClaw](https://openclaw.ai/) 是一个开源的 AI Agent 平台，**在你的设备上运行**，通过你已经在用的聊天应用（WhatsApp、Telegram、Discord、Slack、Teams）来操控 AI 助手。

> **The AI that actually does things.**
>
> 清理收件箱、发送邮件、管理日历、航班值机——全部通过你熟悉的聊天应用完成。

**OpenClaw 核心理念：**

```
Your assistant. Your machine. Your rules.
你的助手。你的设备。你的规则。
```

**与 SaaS 助手的区别：** OpenClaw 运行在你选择的地方——笔记本、家用服务器或 VPS。你的基础设施，你的密钥，你的数据。

**OpenClaw 特点：**
- 🦞 **开源免费** - 100,000+ GitHub Stars
- 🏠 **本地运行** - 数据留在你的设备上
- 💬 **多渠道支持** - WhatsApp、Telegram、Discord、Slack、Teams、Twitch、Google Chat
- 🤖 **多模型支持** - Claude、GPT、DeepSeek、KIMI K2.5、Xiaomi MiMo 等
- 🔌 **ClawHub 技能市场** - 安装和分享 AgentSkills

**官方链接：**
- 官网：[openclaw.ai](https://openclaw.ai/)
- 文档：[docs.openclaw.ai](https://docs.openclaw.ai/)
- 技能市场：[clawhub.ai](https://clawhub.ai/)
- GitHub：[github.com/openclaw/openclaw](https://github.com/openclaw/openclaw)

---

## 安装方式

### 方式一：ClawHub 安装

如果你已安装 `clawhub` CLI，这是最简单的方式：

```bash
# 安装 OpenClaw 专用 md2wechat skill 包
clawhub install md2wechat
```

当前 ClawHub 路径会暴露结构化安装资源；完整、可验证的安装主线仍建议使用下面的固定版本 installer。

**没有 clawhub？先安装它：**

```bash
npm install -g clawhub
clawhub login
```

---

### 方式二：一键脚本安装

适合没有安装 clawhub 的用户：

```bash
export MD2WECHAT_RELEASE_BASE_URL=https://github.com/geekjourneyx/md2wechat-skill/releases/download/v2.0.0
curl -fsSL "${MD2WECHAT_RELEASE_BASE_URL}/install-openclaw.sh" | bash
```

**脚本功能：**
- 按固定版本安装 OpenClaw skill 包与 runtime
- 自动校验 `checksums.txt`
- 安装到 `~/.openclaw/skills/md2wechat/`
- 安装 runtime 到 `~/.openclaw/tools/md2wechat/md2wechat`
- 显示配置说明

---

### 方式三：手动安装

```bash
# 1. 下载固定版本 release 资产
VERSION=2.0.0
# 按你的平台选择对应二进制，这里以 Linux amd64 为例
curl -LO https://github.com/geekjourneyx/md2wechat-skill/releases/download/v${VERSION}/md2wechat-openclaw-skill.tar.gz
curl -LO https://github.com/geekjourneyx/md2wechat-skill/releases/download/v${VERSION}/md2wechat-linux-amd64
curl -LO https://github.com/geekjourneyx/md2wechat-skill/releases/download/v${VERSION}/checksums.txt
sha256sum -c checksums.txt --ignore-missing

# 2. 解压并复制技能目录
mkdir -p /tmp/md2wechat-openclaw
tar -xzf md2wechat-openclaw-skill.tar.gz -C /tmp/md2wechat-openclaw
mkdir -p ~/.openclaw/skills
cp -r /tmp/md2wechat-openclaw/skills/md2wechat ~/.openclaw/skills/

# 3. 安装 runtime
mkdir -p ~/.openclaw/tools/md2wechat
install -m 0755 md2wechat-linux-amd64 ~/.openclaw/tools/md2wechat/md2wechat

# 4. 设置执行权限
chmod +x ~/.openclaw/skills/md2wechat/scripts/*.sh
```

手动安装时，请以同一版本 release 提供的 OpenClaw 资产为准，确保 skill 包与 runtime 同步安装到 OpenClaw 管理目录，不要依赖 `run.sh` 首次运行再下载。

---

## 配置说明

安装完成后，需要配置微信公众号凭证。

### 编辑 OpenClaw 配置文件

打开 `~/.openclaw/openclaw.json`，添加以下配置：

```json
{
  "skills": {
    "entries": {
      "md2wechat": {
        "enabled": true,
        "env": {
          "WECHAT_APPID": "你的AppID",
          "WECHAT_SECRET": "你的Secret"
        }
      }
    }
  }
}
```

### 配置项说明

| 环境变量 | 必需 | 说明 | 获取方式 |
|---------|------|------|---------|
| `WECHAT_APPID` | 草稿上传时 | 微信公众号 AppID | [微信开发者平台](https://developers.weixin.qq.com/platform) → 开发接口管理 |
| `WECHAT_SECRET` | 草稿上传时 | 微信公众号 Secret | 同上，点击"重置"获取 |
| `IMAGE_API_KEY` | AI 图片时 | 图片生成 API Key | 见 [图片服务配置](IMAGE_PROVISIONERS.md) |

### 可选：图片生成配置

如果需要 AI 图片生成功能，添加以下配置：

```json
{
  "skills": {
    "entries": {
      "md2wechat": {
        "enabled": true,
        "env": {
          "WECHAT_APPID": "你的AppID",
          "WECHAT_SECRET": "你的Secret",
          "IMAGE_PROVIDER": "modelscope",
          "IMAGE_API_KEY": "ms-your-token-here",
          "IMAGE_API_BASE": "https://api-inference.modelscope.cn",
          "IMAGE_MODEL": "Tongyi-MAI/Z-Image-Turbo"
        }
      }
    }
  }
}
```

---

## 验证安装

### 检查技能目录

```bash
ls ~/.openclaw/skills/md2wechat/
```

应该看到：
```
SKILL.md
scripts/
references/
```

### 测试运行

```bash
md2wechat --help
```

如果你是在 OpenClaw 管理目录里验证启动器，请确认 runtime 已由安装器安装完成；`run.sh` 只应负责启动，不应在首跑时联网下载。

也可以直接验证 OpenClaw skill 启动器：

```bash
bash ~/.openclaw/skills/md2wechat/scripts/run.sh --help
```

建议再执行一轮发现命令，确认当前 runtime 和资源都可见：

```bash
md2wechat capabilities --json
md2wechat providers list --json
md2wechat themes list --json
md2wechat prompts list --json
```

### 在 OpenClaw 中使用

启动 OpenClaw 后，直接用自然语言交互：

```
请用秋日暖光主题将 article.md 转换为微信公众号格式
```

---

## 常见问题

### Q: 安装后找不到技能？

**A:** 确认技能目录位置正确：

```bash
# 检查目录结构
tree ~/.openclaw/skills/md2wechat/ -L 1
```

如果目录不存在，重新运行安装脚本。

### Q: 运行时报错 "command not found"？

**A:** 检查 runtime 是否已由 OpenClaw 安装器安装完成，并确认 `md2wechat` 可执行文件可用：

```bash
md2wechat --help
```

如果仍然找不到命令，请重新安装 OpenClaw skill 包与 runtime，不要依赖 `run.sh` 首次运行下载。

### Q: 如何更新技能？

**A:**

```bash
# ClawHub 方式
clawhub update md2wechat

# 脚本方式（会覆盖安装）
export MD2WECHAT_RELEASE_BASE_URL=https://github.com/geekjourneyx/md2wechat-skill/releases/download/v2.0.0
curl -fsSL "${MD2WECHAT_RELEASE_BASE_URL}/install-openclaw.sh" | bash
```

### Q: 配置没生效？

**A:** 检查 `openclaw.json` 格式是否正确：

```bash
# 验证 JSON 格式
cat ~/.openclaw/openclaw.json | python3 -m json.tool
```

### Q: 和 Claude Code 安装冲突吗？

**A:** 不冲突。两个平台使用不同的目录：

| 平台 | 技能目录 |
|------|---------|
| Claude Code | `~/.claude/skills/` |
| OpenClaw | `~/.openclaw/skills/` |

可以同时安装在两个平台。

---

## 与 Claude Code 的区别

| 方面 | Claude Code | OpenClaw |
|------|-------------|----------|
| **定位** | 终端 AI 编程助手 | 聊天应用 AI 助手（WhatsApp/Telegram 等） |
| **运行方式** | 本地终端 | 本地运行，通过聊天应用操控 |
| **仓库内 skill 路径** | `skills/md2wechat/` | `platforms/openclaw/md2wechat/` |
| **技能目录** | `~/.claude/skills/` | `~/.openclaw/skills/` |
| **安装方式** | `/plugin` 命令 | `clawhub` CLI / OpenClaw installer |
| **配置文件** | 环境变量 / config.yaml | `openclaw.json` |
| **LLM 支持** | Claude | Claude、GPT、DeepSeek、KIMI 等 |
| **市场** | Plugin Marketplace | [ClawHub](https://clawhub.ai/) |

**说明：** 两个平台共享同一个 CLI 内核，但 skill 包和安装链分开维护。

---

## 相关链接

- [OpenClaw 官网](https://openclaw.ai/)
- [OpenClaw 文档](https://docs.openclaw.ai/)
- [ClawHub 技能市场](https://clawhub.ai/)
- [OpenClaw GitHub](https://github.com/openclaw/openclaw)
- [md2wechat 主仓库](https://github.com/geekjourneyx/md2wechat-skill)
- [问题反馈](https://github.com/geekjourneyx/md2wechat-skill/issues)

---

<div align="center">

**让公众号写作更简单**

</div>
