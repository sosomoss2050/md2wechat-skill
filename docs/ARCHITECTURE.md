# 架构说明

md2wechat-skill 的核心目标不是“把 Markdown 变好看”，而是把文章稳定转换成 **可发布、可验证、可自动化消费** 的微信公众号内容。

## 当前主线

当前代码按这条主线组织：

`cmd -> publish orchestrator -> asset pipeline -> draft/wechat adapters`

这条主线的含义是：

- `cmd/md2wechat` 只负责参数解析、命令入口、输出 envelope 和错误出口。
- `internal/publish` 负责应用层编排，承接文章转换、图片处理、草稿保存和图片帖子创建。
- `internal/publish/AssetPipeline` 负责解析后的资产上传、生成、下载和 HTML 回填。
- `internal/draft` 和 `internal/wechat` 负责平台适配，不再承担命令级业务编排。

## 平台适配层

仓库把 skill 封装分成两条路径，避免把不同平台的安装模型混在一起：

- `skills/md2wechat/`：面向 Claude Code / Codex / OpenCode 的 coding-agent skill。
- `platforms/openclaw/md2wechat/`：面向 OpenClaw / ClawHub 的专用 skill 包。

两条路径共享同一个 CLI 内核，但平台适配不同：

- coding-agent skill 可以保留面向终端工作流的自然语言说明。
- OpenClaw skill 需要结构化声明依赖、安装和凭证要求。
- OpenClaw 安装主线是 skill 包与 runtime 一起安装，`run.sh` 只负责启动，不应在首跑时动态下载二进制。

## 模块职责

### `cmd/md2wechat`

- Cobra 命令定义
- 参数校验
- 统一 JSON envelope
- CLI 错误码与退出路径

### `internal/publish`

- `Service`: 图文发布主编排
- `ImagePostService`: 图片帖子主编排
- `AssetPipeline`: 发布期图片处理与回填
- `model.go`: canonical article / asset / artifact model

### `internal/converter`

- Markdown 转 HTML
- frontmatter / metadata 提取
- Markdown 图片引用解析
- AI prompt 构建与 API/AI 转换模式

### `internal/image`

- 图片生成
- 图片压缩
- 上传/下载能力的运行时注入

### `internal/draft`

- 标准图文草稿 adapter
- 图片帖子 `newspic` adapter
- 摘要生成与微信草稿结构映射

### `internal/wechat`

- 微信 SDK 封装
- 素材上传与重试
- 新建草稿 / newspic draft
- 远程下载边界与 SSRF 防护

## 两条发布流

### `convert`

1. 读取 Markdown
2. 提取 metadata
3. 解析图片引用
4. 执行 API 或 AI 转换
5. 必要时通过 `AssetPipeline` 上传并回填 HTML
6. 可选保存 draft JSON
7. 可选上传封面并创建微信草稿

### `create_image_post`

1. 读取标题、内容和图片输入
2. 必要时从 Markdown 提取本地图片
3. 通过 `ImagePostService` 统一归一化为 `AssetRef`
4. 通过 `AssetPipeline` 上传图片
5. 通过 `draft` / `wechat` adapter 创建 `newspic` 草稿

## 稳定契约

当前工程里有 4 个关键契约：

1. `internal/publish/model.go`
   - 统一 article / asset / artifact model
2. `internal/action/action.go`
   - 统一 `completed / action_required / failed`
3. CLI JSON envelope
   - 统一 `success / code / message / schema_version / status / retryable / data / error`
4. `AssetPipeline`
   - 统一图片上传、生成、下载、回填

## 当前设计原则

- 命令层不做业务编排
- 平台适配层不做内容预处理
- 图片链只保留一套发布期实现
- 机器输出优先稳定，再考虑人类可读文案
- 文档只描述已经落地、已验证的路径

## 下一阶段建议

当前本地架构已经基本收口。下一阶段优先级不再是继续拆内部模块，而是：

1. 做真实微信/API 凭证的 smoke/staging 验证
2. 继续收口外部系统契约，而不是继续刷局部覆盖率
3. 只有在出现第二个发布后端时，才考虑把 `AssetPipeline` 再抽成独立 domain
