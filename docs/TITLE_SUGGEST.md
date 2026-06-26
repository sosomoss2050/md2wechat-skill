# 公众号标题建议

这篇文档讲清楚 `md2wechat title suggest` 怎么用、它会输出什么、`--hook-level` 如何影响提示词，以及 Agent / 人类用户应该如何完成从文章到最终标题的闭环。

核心结论先放前面：

- `title suggest` 不是本地标题生成器。
- 它会读取文章，渲染内置标题 prompt，然后把 prompt 放进 JSON 返回。
- 真正生成标题候选的是宿主 Agent 或外部模型。
- CLI 不调用模型，不自动选最终标题，不写回 Markdown，不创建微信草稿。
- 最终标题应该由用户确认，或由上层 Agent 基于候选评分推荐后再确认。

## 适用场景

适合使用：

- 已有一篇 Markdown 文章，需要生成 8-10 个公众号标题候选。
- 想让 Agent 根据文章内容生成标题，但不希望 Agent 直接改文章。
- 想比较克制标题、强一点的标题、反差更强的标题。
- 想保留事实边界，避免“标题党”变成虚假承诺。

不适合使用：

- 你只想手动覆盖最终草稿标题。这个场景直接用 `convert --title` 或 frontmatter `title`。
- 你希望 CLI 直接调用模型生成标题。当前 CLI 只准备请求，不调用模型。
- 你希望 CLI 自动选择最终标题并写回文章。当前没有这个副作用。

## 最小命令

```bash
md2wechat title suggest article.md --json
```

返回的是一个 JSON envelope。成功时核心字段是：

```json
{
  "success": true,
  "code": "TITLE_SUGGEST_REQUEST_READY",
  "status": "action_required",
  "data": {
    "action": "ai_title_suggestion_request",
    "execution_owner": "host_agent",
    "prompt_kind": "title",
    "prompt_name": "wechat-title-expert",
    "prompt": "...已经渲染好的标题生成提示词...",
    "hook_level": 1,
    "hook_level_label": "restrained",
    "side_effects": false,
    "requires_external_model": true,
    "recommendation_only": true
  }
}
```

看到 `status: "action_required"` 的意思是：CLI 已经准备好请求，下一步需要 Agent 或外部模型继续执行 `data.prompt`。

## 完整流程

```text
Markdown article.md
   |
   | 1. 用户或 Agent 运行
   |    md2wechat title suggest article.md --json
   v
md2wechat CLI
   |
   | 2. 读取 Markdown
   | 3. 解析 frontmatter title 和正文内容
   | 4. 校验参数
   | 5. 渲染内置 prompt: wechat-title-expert
   v
JSON envelope
   |
   | code: TITLE_SUGGEST_REQUEST_READY
   | status: action_required
   | data.prompt: 给模型执行的完整提示词
   | side_effects: false
   | requires_external_model: true
   | recommendation_only: true
   v
Host Agent / External Model
   |
   | 6. 执行 data.prompt
   | 7. 返回标题候选 JSON
   | 8. 每个候选包含评分、证据依据、风险标记
   v
User / Upper Agent
   |
   | 9. 用户确认最终标题
   |    或 Agent 基于 weighted_score 推荐一个标题给用户确认
   v
Final title
   |
   | 10. 后续再进入 convert / preview / create_draft 等流程
   v
Publish workflow
```

这条链路中，`title suggest` 只负责 1-5 步。后面的模型生成、标题确认和发布动作都不属于这个命令。

## 常用参数

```bash
md2wechat title suggest article.md \
  --target-reader "独立开发者" \
  --count 10 \
  --max-title-chars 25 \
  --hook-level 2 \
  --json
```

参数说明：

| 参数 | 默认值 | 范围 | 作用 |
|------|--------|------|------|
| `--target-reader` | 从文章推断 | 任意文本 | 告诉模型标题要打给谁 |
| `--count` | `10` | `8..10` | 希望生成多少个候选标题 |
| `--max-title-chars` | `25` | `12..32` | 单个标题最长字数 |
| `--hook-level` | `1` | `1..3` | 标题钩子张力 |
| `--prompt` | `wechat-title-expert` | 已存在 title prompt | 使用哪个标题 prompt |
| `--json` | 必填 | - | 输出机器可读 JSON |

`title suggest` 必须带 `--json`。这是为了让 Agent 能稳定读取 `data.prompt`、`side_effects`、`requires_external_model` 等字段。

## hook-level 是什么

`--hook-level` 是标题张力控制变量。

它不是：

- 不是标题党开关。
- 不是 prompt family 切换。
- 不是本地模型选择。
- 不是自动最终标题选择。

它只是把两个变量注入同一个 prompt：

```text
{{HOOK_LEVEL}}
{{HOOK_LEVEL_LABEL}}
```

例如：

```bash
md2wechat title suggest article.md --json --hook-level 2
```

会渲染成：

```text
标题钩子力度：2（punchy；1=restrained，2=punchy，3=high_tension）
```

三个 level 的含义：

| Level | Label | 适合场景 | 风险 |
|-------|-------|----------|------|
| `1` | `restrained` | 默认、稳健、可信、偏直给价值 | 点击欲望可能不够强 |
| `2` | `punchy` | 想增强反差、数字、实体、后果感 | 需要确保文章内容支撑 |
| `3` | `high_tension` | 高张力标题实验、强反差选题 | 最容易越界，必须有证据和风险标记 |

推荐使用方式：

```text
普通文章 / 默认生成          -> --hook-level 1
希望更有传播性但仍稳健      -> --hook-level 2
明确做高张力标题实验        -> --hook-level 3
```

多数情况下建议从 `--hook-level 2` 开始，而不是直接用 3。

## hook-level 如何影响提示词

`--hook-level` 最终会影响 prompt 中这几段：

```text
Level 1 restrained：
保持当前克制、可信、内容价值优先的标题力度。

Level 2 punchy：
在文章支持时，可以使用更强对比、标点、明确实体、数字和后果框架，但不得扩大事实。

Level 3 high_tension：
可以尝试高张力表达，但每个候选都必须有证据依据和风险披露；如果证据不足，必须降低标题力度，并在 recommendation_reason 中说明。
```

也会影响模型输出 JSON 中的字段：

```json
{
  "hook_level": 2,
  "hook_level_label": "punchy"
}
```

注意：CLI 不会根据 level 生成不同 prompt 文件。当前只有一个内置标题 prompt：`wechat-title-expert`。

## 事实边界

标题可以更有吸引力，但不能编造事实。

这些词有明确使用边界：

| 表达 | 什么时候可以用 | 什么时候不要用 |
|------|----------------|----------------|
| `刚刚` | 文章包含明确当前时间信号 | 只是普通复盘、旧闻、无时间证据 |
| `全网` | 文章包含传播范围、平台范围、讨论量证据 | 没有传播范围证据 |
| `第一` | 文章包含排名证据 | 只是作者主观判断 |
| `榜首` | 文章包含榜单或排名来源 | 没有榜单 |
| `变天` | 文章能支撑行业变化或竞争格局变化 | 只是小功能、小更新 |

如果文章证据不足，模型应该降低标题张力，并在 `recommendation_reason` 里说明原因。

## 模型应该返回什么

`data.prompt` 要求外部模型返回 strict JSON，不要返回 Markdown 表格。

候选标题结构大致是：

```json
{
  "article_summary": "文章讲什么",
  "core_value": "读者读完能得到什么",
  "target_reader": "目标读者",
  "title_count": 10,
  "max_title_chars": 25,
  "hook_level": 2,
  "hook_level_label": "punchy",
  "candidates": [
    {
      "rank": 1,
      "title": "候选标题",
      "strategy": "direct_value",
      "claim_strength": "medium",
      "evidence_basis": "标题承诺对应的文章依据",
      "risk_flags": ["none"],
      "scores": {
        "curiosity": 8,
        "value": 9,
        "truthfulness": 9,
        "rhythm": 8,
        "spreadability": 8
      },
      "weighted_score": 8.45,
      "recommendation_reason": "为什么推荐，以及为什么没有超过文章内容"
    }
  ]
}
```

评分公式：

```text
weighted_score =
  curiosity * 0.25
  + value * 0.25
  + truthfulness * 0.20
  + rhythm * 0.15
  + spreadability * 0.15
```

字段含义：

| 字段 | 含义 |
|------|------|
| `curiosity` | 好奇缺口 |
| `value` | 价值预期 |
| `truthfulness` | 标题承诺是否被文章支撑 |
| `rhythm` | 中文节奏和口语顺滑度 |
| `spreadability` | 转发讨论潜力 |
| `claim_strength` | 标题承诺强度 |
| `evidence_basis` | 标题关键承诺的文章依据 |
| `risk_flags` | 是否存在事实边界风险 |

## 人类用户怎么用

### 第一步：先检查文章状态

```bash
md2wechat inspect article.md --json
```

确认标题、摘要、H1、图片等基础信息没有明显问题。

### 第二步：生成标题建议请求

稳健版：

```bash
md2wechat title suggest article.md --json --hook-level 1
```

推荐版：

```bash
md2wechat title suggest article.md --json --hook-level 2
```

高张力实验版：

```bash
md2wechat title suggest article.md --json --hook-level 3
```

### 第三步：把 `data.prompt` 给模型

如果你在 Agent 环境里，Agent 应该自动读取 `data.prompt` 并交给外部模型。

如果你手动操作，可以把 JSON 里的 `data.prompt` 复制给你使用的模型，让它返回候选 JSON。

### 第四步：选标题

推荐判断顺序：

```text
1. 先看 truthfulness
2. 再看 value
3. 再看 curiosity
4. 最后看 rhythm / spreadability
```

不要只看总分。一个标题如果 `truthfulness` 低，就算 `curiosity` 高，也不应该作为最终标题。

### 第五步：确认最终标题后再发布

最终标题确认后，再进入常规流程：

```bash
md2wechat inspect article.md --json
md2wechat preview article.md --json
md2wechat convert article.md --title "最终标题"
```

或者在创建草稿时显式覆盖标题。

## Agent 怎么用

Agent 推荐流程：

```text
1. md2wechat capabilities --json
   |
   | 确认 data.title_generation.available = true
   v
2. md2wechat title suggest article.md --json --hook-level <1|2|3>
   |
   | 读取 data.prompt
   | 确认 side_effects=false
   | 确认 requires_external_model=true
   | 确认 recommendation_only=true
   v
3. 调用宿主模型执行 data.prompt
   |
   | 得到 candidates[]
   v
4. 如果用户要求确认：
      展示候选并让用户选

   如果用户要求自动推荐：
      只推荐 weighted_score 高且 truthfulness 足够高的标题
      不直接写回文章
   v
5. 用户确认后，再进入 convert / draft 流程
```

Agent 不应该做：

- 不要把 `TITLE_SUGGEST_REQUEST_READY` 当成标题已经生成。
- 不要把第 1 名候选自动写回 Markdown。
- 不要自动创建草稿。
- 不要把 `--hook-level 3` 当成默认值。
- 不要把 `risk_flags` 非 `none` 的标题直接当最终标题。

## 错误处理

常见错误：

| 错误码 | 原因 | 处理方式 |
|--------|------|----------|
| `CONFIG_INVALID` | 没有传 `--json` | 加上 `--json` |
| `TITLE_SUGGEST_READ_FAILED` | 文章文件不存在或无法读取 | 检查路径 |
| `TITLE_SUGGEST_INVALID` | 参数非法 | 检查 `--count`、`--max-title-chars`、`--hook-level` |

`--hook-level` 只接受数字：

```bash
md2wechat title suggest article.md --json --hook-level 2
```

这些都会失败：

```bash
md2wechat title suggest article.md --json --hook-level punchy
md2wechat title suggest article.md --json --hook-level explosive
md2wechat title suggest article.md --json --hook-level 4
md2wechat title suggest article.md --json --hook-level=
```

失败时会返回：

```json
{
  "success": false,
  "code": "TITLE_SUGGEST_INVALID",
  "status": "failed"
}
```

## 如何查看当前内置 prompt

查看 prompt 元数据：

```bash
md2wechat prompts show wechat-title-expert --kind title --json
```

渲染 prompt：

```bash
md2wechat prompts render wechat-title-expert \
  --kind title \
  --var ARTICLE_CONTENT="$(cat article.md)" \
  --var TARGET_READER="独立开发者" \
  --var TITLE_COUNT=10 \
  --var MAX_TITLE_CHARS=25 \
  --var HOOK_LEVEL=2 \
  --var HOOK_LEVEL_LABEL=punchy \
  --json
```

更推荐直接使用：

```bash
md2wechat title suggest article.md --json --hook-level 2
```

因为 `title suggest` 会帮你处理 Markdown 解析、默认值、参数校验和 JSON envelope。

## 和文章最终标题的关系

`title suggest` 只生成标题候选请求，不改变文章最终标题。

文章最终标题仍由常规元数据规则决定：

```text
--title
  -> frontmatter.title
  -> 正文首个 Markdown H1
  -> 未命名文章
```

如果你确认了一个最终标题，有两种常见做法：

```bash
md2wechat convert article.md --title "最终标题"
```

或者修改 frontmatter：

```markdown
---
title: 最终标题
---
```

## 推荐实践

默认推荐：

```bash
md2wechat title suggest article.md --json --hook-level 2
```

保守场景：

```bash
md2wechat title suggest article.md --json --hook-level 1
```

高张力实验：

```bash
md2wechat title suggest article.md --json --hook-level 3
```

高张力标题必须满足：

- 有明确文章证据。
- `evidence_basis` 非空。
- `risk_flags` 存在。
- 不编造数字、实体、排名、传播范围或当前时间。
- 用户确认后才进入发布流程。

## 一句话记忆

```text
title suggest 负责准备标题生成请求；
外部模型负责生成候选；
用户或上层 Agent 负责确认最终标题；
发布命令才负责进入发布链路。
```
