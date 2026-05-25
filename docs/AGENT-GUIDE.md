# md2wechat Agent 操作手册

本文档为 AI Agent 提供操作 md2wechat CLI 的完整指引，涵盖所有典型场景的命令序列。

> **阅读对象**：Claude Code、OpenClaw、Copilot CLI 等 AI 代理（Agent），以及希望理解 Agent 决策流程的用户。

---

## 一、能力发现（每次会话开始前执行）

在执行任何排版任务前，Agent 必须先发现当前实例的能力边界。这些命令是**幂等的**，应当在每个新会话的开始时执行一次：

### 1.1 查看完整能力集

```bash
md2wechat capabilities --json
```

响应内容包括：
- 已开放的所有高层命令
- `convert` 支持的模式（api / ai）与默认模式
- `inspect` 和 `preview` 的可用选项
- 支持的图片 provider、主题、prompt 种类数量

### 1.2 查看可用主题（38+ 个）

```bash
md2wechat themes list --json
```

列出所有当前可用的主题。主题遵循以下加载优先级：
1. `MD2WECHAT_THEMES_DIR` 环境变量指向的目录
2. `./themes` 本地目录
3. `~/.config/md2wechat/themes` 用户目录
4. 内置主题资产（38 个）

使用 `md2wechat themes show <theme_name> --json` 查看单个主题的详细配置。

### 1.3 查看可用图片 Provider（6+ 个）

```bash
md2wechat providers list --json
```

列出所有支持的图片生成 provider：
- `openai` — OpenAI GPT Image / DALL·E
- `tuzi` — 图子 AI
- `modelscope` / `ms` — ModelScope
- `openrouter` / `or` — OpenRouter
- `gemini` / `google` — Google Gemini
- `volcengine` / `volc` — 火山引擎

使用 `md2wechat providers show <provider_name> --json` 查看单个 provider 需要的配置和支持的模型。

### 1.4 查看 Prompt 模板库

```bash
md2wechat prompts list --json
```

列出所有内置 prompt 模板，包括：
- `humanizer` — 人性化润色（4 个强度 + authentic 模式）
- `refine` — 内容精修
- `image` — 图片生成

使用 `md2wechat prompts show <name> --kind <kind> --json` 查看单个 prompt 的具体内容和配置。

### 1.5 查看高级排版模块（43 个内置模块）

```bash
md2wechat layout list --json
md2wechat layout list --serves attention --json     # 按目标筛选
md2wechat layout list --serves readability --json
md2wechat layout list --serves memorability --json
md2wechat layout list --serves conversion --json
```

列出所有可用的高级排版模块。每个模块都有四个服务目标之一或多个：
- `attention` — 吸引注意力（hero、verdict、toc）
- `readability` — 提升可读性（steps、callout、bridge）
- `memorability` — 增强记忆点（quote、metrics、summary）
- `conversion` — 驱动转化（cta、subscribe、faq）

**⚠️ 重要约束**：高级排版语法（`:::module` ... `:::` 语法）**仅在 API 模式**下渲染。使用 `--mode ai` 时，排版语法会被当作普通段落输出。

---

## 二、配置发现与验证

### 2.1 查看当前配置

```bash
md2wechat config show --format json
```

这个命令返回当前生效的配置，已合并以下三层优先级：
1. 配置文件 `~/.config/md2wechat/config.yaml`（或通过 `--config` 指定）
2. 环境变量（如 `WECHAT_APPID`、`IMAGE_API_KEY`）
3. 硬编码默认值

Agent 应该用这个命令来确认当前配置是否就绪，而不是手动逐一检查。

### 2.2 初始化配置文件（仅用户第一次）

```bash
md2wechat config init
```

这会在 `~/.config/md2wechat/config.yaml` 创建示例配置文件。Agent 不应该自动调用此命令，除非用户明确要求重新初始化。

### 2.3 验证配置（发布前必须）

```bash
md2wechat config validate
```

在发布任何文章到微信前，都应该先验证配置。如果返回非零退出码，表示配置有问题。

---

## 三、Brand Profile 操作

Brand Profile 是一个可选的品牌档案，位于 `~/.config/md2wechat/brand.md`，由 Agent 读取但不被 CLI 解析。

它使用 **Markdown 格式**，而不是 YAML，让 Agent 可以直接理解自然语言描述的品牌风格和偏好。

### 3.1 检查 Brand Profile 是否存在

```bash
md2wechat brand show --json
```

响应解读：
- `code: "BRAND_SHOWN"` → 档案存在，`data.content` 包含完整 Markdown 文本
- `code: "BRAND_NOT_FOUND"` → 档案不存在，可引导初始化
- `code: "BRAND_READ_FAILED"` → 文件不可读（权限问题等）

### 3.2 读取 Brand Profile

Agent 读取方式：

```python
import os

brand_path = os.path.expanduser("~/.config/md2wechat/brand.md")
brand_content = ""
if os.path.exists(brand_path):
    with open(brand_path) as f:
        brand_content = f.read()

# brand_content 包含完整的 Markdown 文本
# Agent 将其作为上下文注入到排版决策中
```

或通过 CLI 命令获取：

```bash
result=$(md2wechat brand show --json)
brand_content=$(echo "$result" | jq -r '.data.content')
```

### 3.3 Brand Profile 引导流程（如果不存在）

当 `md2wechat brand show` 返回 `BRAND_NOT_FOUND` 时，Agent 应发起 3 问引导：

```
我注意到你还没有设置品牌档案。这能让每篇文章都"像你说的话"。
3 个问题，2 分钟完成：

1. 你希望文章的整体语气是？（例如：犀利实用、温暖鼓励、专业严谨）
2. 有没有你一定要避免的表达方式？（例如：过多 emoji、空泛鸡汤）
3. 如果你有风格参考文件或文件夹，可以告诉我路径（可选）
```

根据用户回答，创建或编辑 `~/.config/md2wechat/brand.md`。

---

## 四、Markdown 排版诊断

在将文章交给 `convert` 之前，Agent 应该先做"排版诊断"来理解文章的结构和需求。

### 4.1 诊断三步法

**Step 1 — 意图识别**

Agent 应该提出三个问题来理解文章：
- 这篇文章的目标读者是谁？
- 这篇文章想让读者下一步做什么（conversion goal）？
- 读完后读者最应该记住什么（memorability anchor）？

**Step 2 — 内容映射**（基于四目标框架）

使用 43 个排版模块中的这些映射关系：

| 目标 | 常见问题 | 推荐模块 | 示例 |
|------|---------|---------|------|
| **attention** | 这篇值不值得看？ | hero, verdict, toc, audience-fit | 用 hero 展示视觉吸引力 |
| **readability** | 手机上读起来累不累？ | steps, callout, bridge, part | 用 callout 突出关键信息 |
| **memorability** | 读完记住什么？ | quote, metrics, summary, author-card | 用 metrics 展示数据 |
| **conversion** | 下一步做什么？ | cta, subscribe, faq, notice | 用 cta 明确行动 |

**Step 3 — 模块选择与验证**

结合 Brand Profile 的 `limits` 和内容特性，从 43 个模块中选择最合适的组合，然后验证语法：

```bash
md2wechat layout validate --file article.md --json
```

### 4.2 查看单个模块详情与示例

```bash
# 查看模块完整规范
md2wechat layout show hero --json
md2wechat layout show callout --json

# 查看渲染示例（需要 API 服务）
md2wechat layout render callout --var CONTENT="重要提示" --var LEVEL="warning"
```

---

## 五、文章转换（convert）

### 5.1 基础转换命令

#### 模式 A：API 模式（默认，支持高级排版）

```bash
md2wechat convert article.md --output output.html
```

**前置条件**：
- 需要本地 API 服务运行在 `http://localhost:3000`（或通过 `API_BASE_URL` 配置）
- 支持所有 43 个高级排版模块
- 输出 HTML 最终，可直接用于微信草稿或发布

**响应示例**：
```json
{
  "success": true,
  "code": "CONVERTED",
  "data": {
    "output_file": "/path/to/output.html",
    "html_size": 12345,
    "images_count": 3,
    "layout_modules_used": ["hero", "callout", "cta"]
  }
}
```

#### 模式 B：AI 模式（轻量级，不需要 API）

```bash
md2wechat convert article.md --mode ai --output output.html
```

**特点**：
- 不需要本地 API 服务
- 不支持高级排版语法（`:::module` 被当作普通文本）
- 适合快速预览或离线场景

**⚠️ 降级规则**：如果 API 服务不可用，自动降级到 AI 模式。

#### 模式 C：指定主题

```bash
md2wechat convert article.md --theme bytedance --output output.html
md2wechat convert article.md --theme minimal-blue --output output.html
md2wechat convert article.md --theme autumn-warm --output output.html
```

#### 模式 D：指定其他参数

```bash
md2wechat convert article.md \
  --title "自定义标题" \
  --author "自定义作者" \
  --digest "自定义摘要" \
  --output output.html
```

### 5.2 验证文章元数据（发布前必须）

```bash
md2wechat inspect article.md --json
```

响应包含：
- `readiness` — 发布就绪状态
- `checks` — 详细检查项（如 `TITLE_TOO_LONG`, `MISSING_COVER`）
- `metadata` — 解析出的标题、作者、摘要、图片数量

Agent 应检查 `readiness.ready_for_draft` 或 `readiness.ready_for_upload` 来决定是否可以继续。

### 5.3 预览文章

```bash
md2wechat preview article.md
md2wechat preview article.md --output /path/to/preview.html
```

在浏览器中打开预览，Agent 等待用户确认后才执行后续操作。

---

## 六、发布流程

### 6.1 检查图片是否需要上传

```bash
md2wechat inspect article.md --upload --json
```

检查 `data.readiness.ready_for_upload` 字段，确认是否满足上传条件。

### 6.2 上传封面图

```bash
md2wechat generate_cover "article title" --output cover.jpg
md2wechat upload_image cover.jpg --json
```

响应包含 `media_id`，记录此 ID 用于后续发布。

### 6.3 创建草稿

```bash
md2wechat convert article.md --draft --cover ./cover.jpg --output draft.html
md2wechat create_draft draft.html --json
```

也可以一步到位：

```bash
md2wechat convert article.md \
  --draft \
  --cover ./cover.jpg \
  --output draft.html
```

响应包含 `media_id` 和 `publish_url`，用于后续发布。

### 6.4 创建图文消息（Image Post / 小绿书）

```bash
md2wechat create_image_post article.md --cover ./cover.jpg --json
```

适用于内容更适合以卡片形式展示的场景。

---

## 七、高级排版模块操作

### 7.1 验证排版语法

在转换前，始终验证 Markdown 文件中的排版语法：

```bash
md2wechat layout validate --file article.md --json
```

响应解读：
- `code: "LAYOUT_VALIDATED"` → 语法正确，可以转换
- `code: "LAYOUT_VALIDATE_ERRORS"` → 有错误，查看 `data.errors` 了解详情
- `data.errors` — 包含错误位置和修复建议

### 7.2 在文章中使用排版模块

在 Markdown 文件中使用 `:::` 语法（**仅 API 模式下生效**）：

```markdown
:::hero
title: 这是标题
subtitle: 这是副标题
:::

:::verdict
content: 这篇文章的核心观点
confidence: high
:::

:::callout level=warning
这是一个重要提示
:::

:::quote
content: 经典引用
author: 引用来源
:::

:::cta
title: 如果这篇对你有启发
action: 关注 / 分享 / 咨询
:::
```

### 7.3 查看模块完整规范

```bash
md2wechat layout show hero --json
md2wechat layout show verdict --json
md2wechat layout show callout --json
```

### 7.4 测试模块渲染

```bash
md2wechat layout render callout --var CONTENT="测试内容" --var LEVEL="info"
md2wechat layout render cta --var TITLE="立即订阅" --var ACTION="subscribe"
```

---

## 八、异常处理与降级

### 8.1 本地 API 服务未启动

**症状**：`convert` 命令返回 `connection refused` 错误

**处理**：
```bash
# 1. 检查本地服务状态
curl -s http://localhost:3000/ || echo "❌ 服务未启动"

# 2. 降级到 AI 模式
md2wechat convert article.md --mode ai --output output.html

# 3. 或者启动本地服务（由用户操作）
# 用户应该启动 md2wechat API 服务
```

### 8.2 Brand Profile 不存在

**症状**：`md2wechat brand show` 返回 `BRAND_NOT_FOUND`（当命令可用时）

**处理**：
```bash
# 1. 发起 3 问流程，引导用户设置品牌档案
# Agent: 我注意到你还没有设置品牌档案。这能让每篇文章都"像你说的话"。
#       3 个问题，2 分钟完成：
#       1. 你希望文章的整体语气是？（例如：犀利实用、温暖鼓励、专业严谨）
#       2. 有没有你一定要避免的表达方式？（例如：过多 emoji、空泛鸡汤）
#       3. 如果你有风格参考文件或文件夹，可以告诉我路径（可选）

根据用户回答，创建或编辑 `~/.config/md2wechat/brand.md`。
# 或等待 md2wechat brand init 命令上线
```

### 8.3 排版语法错误

**症状**：`md2wechat layout validate --file article.md --json` 返回错误

**处理**：
```bash
# 1. 查看具体错误
md2wechat layout validate --file article.md --json | jq '.data.errors'

# 2. 查看对应模块的正确语法
md2wechat layout show <failing_module_name> --json

# 3. 查看实际渲染示例
md2wechat layout render <module_name> --var KEY=VALUE

# 4. 修复后重新验证
md2wechat layout validate --file article.md --json
```

### 8.4 发布失败

常见原因及处理：

| 错误 | 原因 | 处理方法 |
|------|------|---------|
| `MISSING_COVER` | 缺少封面图 | 上传封面图后重试 |
| `INVALID_TOKEN` | 微信凭证过期 | 检查 `config show --format json` 中的 token，重新授权 |
| `TITLE_TOO_LONG` | 标题超过 64 字符 | 修改标题后重试 |
| `IMAGE_REPLACEMENT_REQUIRES_UPLOAD` | 图片需要上传到微信 | 使用 `--upload` 参数或提前上传 |
| `ip_not_in_whitelist` | 微信 IP 白名单未配置 | 参考 [WECHAT-CREDENTIALS.md](WECHAT-CREDENTIALS.md) |

---

## 九、决策优先级链

当面临多个选择时，按以下顺序应用决策：

```
1. CLI flag（如 --theme default）
    ↓
2. Brand Profile（brand.md 中的语气风格、排版约束）
    ↓
3. 环境变量（如 CONVERT_MODE=ai）
    ↓
4. 配置文件（~/.config/md2wechat/config.yaml）
    ↓
5. 排版诊断推荐（基于 43 模块的四目标框架）
    ↓
6. 硬编码默认值（主题 default，模式 api）
```

---

## 十、完整场景示例

### 场景 A：用户首次发文，无配置无 Brand Profile

```bash
# 1. 发现能力（必须）
md2wechat capabilities --json
md2wechat themes list --json
md2wechat layout list --json

# 2. 检查并初始化配置
md2wechat config show --format json
# 如果缺少关键配置，提示用户运行 md2wechat config init

# 3. 发起 3 问流程创建 Brand Profile（当命令可用时）
# 或手动创建 ~/.config/md2wechat/brand.md

# 4. 验证排版语法
md2wechat layout validate --file article.md --json

# 5. 转换
md2wechat convert article.md --output output.html

# 6. 预览
md2wechat preview article.md

# 7. 检查发布就绪
md2wechat inspect article.md --draft --cover ./cover.jpg --json

# 8. 创建草稿
md2wechat convert article.md \
  --draft \
  --cover ./cover.jpg \
  --output draft.html
```

### 场景 B：用户有完整配置，快速发文

```bash
# 1. 验证排版语法
md2wechat layout validate --file article.md --json

# 2. 转换（复用当前主题和配置）
md2wechat convert article.md --output output.html

# 3. 预览
md2wechat preview article.md

# 4. 一键创建草稿和发布
md2wechat convert article.md \
  --draft \
  --cover ./cover.jpg \
  --theme bytedance
```

### 场景 C：调试排版问题

```bash
# 1. 验证语法并查看具体错误
md2wechat layout validate --file article.md --json

# 2. 查看失败模块的规范
md2wechat layout show <failing_module> --json

# 3. 查看模块渲染示例
md2wechat layout render <failing_module> --var KEY=VALUE

# 4. 修复文件中的语法错误

# 5. 重新验证
md2wechat layout validate --file article.md --json

# 6. 转换并预览
md2wechat convert article.md --mode api --output output.html
md2wechat preview article.md --output preview.html
```

### 场景 D：切换主题或 Provider

```bash
# 1. 查看可用主题
md2wechat themes list --json

# 2. 使用新主题转换
md2wechat convert article.md --theme minimal-blue --output output.html

# 3. 查看可用 Provider
md2wechat providers list --json

# 4. 更新配置或环境变量以切换 Provider
export IMAGE_PROVIDER=tuzi
export IMAGE_API_KEY=your_api_key

# 5. 重新转换
md2wechat convert article.md --output output.html
```

### 场景 E：处理 API 服务不可用

```bash
# 1. 检查 API 服务
curl -s http://localhost:3000/ || echo "服务未启动"

# 2. 降级到 AI 模式
md2wechat convert article.md --mode ai --output output.html

# 3. 注意：AI 模式不支持高级排版模块
# 预期行为：:::module ... ::: 语法被当作普通文本

# 4. 如需使用排版模块，必须启动 API 服务后使用 API 模式
```

---

## 十一、常用命令快速参考

| 目标 | 命令 |
|------|------|
| 查看当前能力 | `md2wechat capabilities --json` |
| 查看可用主题 | `md2wechat themes list --json` |
| 查看可用 Provider | `md2wechat providers list --json` |
| 查看 Prompt 模板 | `md2wechat prompts list --json` |
| 查看排版模块 | `md2wechat layout list --json` |
| 验证排版语法 | `md2wechat layout validate --file article.md --json` |
| 查看配置 | `md2wechat config show --format json` |
| 验证配置 | `md2wechat config validate` |
| 转换文章（API） | `md2wechat convert article.md --output output.html` |
| 转换文章（AI） | `md2wechat convert article.md --mode ai --output output.html` |
| 检查发布就绪 | `md2wechat inspect article.md --draft --json` |
| 预览文章 | `md2wechat preview article.md` |
| 创建草稿 | `md2wechat convert article.md --draft --cover ./cover.jpg` |
| 查看模块规范 | `md2wechat layout show hero --json` |
| 测试模块渲染 | `md2wechat layout render callout --var CONTENT="test" --var LEVEL="info"` |

---

## 十二、调试技巧

### 启用 JSON 输出用于脚本集成

几乎所有命令都支持 `--json` 参数，返回机器可读的 JSON envelope：

```bash
md2wechat convert article.md --json
md2wechat inspect article.md --json
md2wechat layout validate --file article.md --json
```

JSON envelope 格式（v1）：
```json
{
  "success": true,
  "code": "COMMAND_CODE",
  "message": "human-readable message",
  "schema_version": "v1",
  "status": "completed|action_required|failed",
  "retryable": false,
  "data": {},
  "error": ""
}
```

### 检查退出码

- `0` — 成功
- `1` — 通用错误
- `2` — 校验失败（使用 `--strict` 时可获得此码）

### 常见的 JSON 响应码

| code | 含义 | 处理方式 |
|------|------|---------|
| `CONVERTED` | 转换成功 | 继续发布流程 |
| `INSPECTED` | 检查完成 | 查看 `data.readiness` |
| `LAYOUT_VALIDATED` | 排版语法正确 | 可以转换 |
| `LAYOUT_VALIDATE_ERRORS` | 排版有错误 | 查看 `data.errors` 修复 |
| `BRAND_NOT_FOUND` | Brand Profile 不存在 | 可选初始化 |
| `BRAND_SHOWN` | Brand Profile 已读取 | 使用其中配置 |
| `CONFIG_SHOWN` | 配置已读取 | 查看 `data` |

---

*最后更新：与 md2wechat v2.2.0 同步*
*下一计划版本 (v2.3)：md2wechat brand 命令*
