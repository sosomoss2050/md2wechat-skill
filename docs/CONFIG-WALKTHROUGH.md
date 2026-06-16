# md2wechat 配置保姆级指南

这份文档按操作顺序解释配置流程。它适合第一次配置、迁移到多公众号、或者让 Agent 排查配置问题。

如果你只想查字段含义，去看 [配置指南](CONFIG.md)。如果你还没有微信公众号 `AppID` / `AppSecret`，先看 [微信凭证与 IP 白名单指南](WECHAT-CREDENTIALS.md)。

---

## 先判断你要做什么

不同任务需要的凭证不同：

| 目标 | 需要配置 |
|------|----------|
| 只把 Markdown 转 HTML | `api.md2wechat_key` |
| 使用高级排版模块 | `api.md2wechat_key`，并使用 API 模式 |
| 上传图片到微信 | `wechat.appid` / `wechat.secret` |
| 创建微信草稿 | `wechat.appid` / `wechat.secret`，草稿封面 |
| 管理多个公众号 | `wechat.accounts`，并为命名账号副作用配置有效的 `MD2WECHAT_API_KEY` |

本地只读命令不会联网验证 API key：

```bash
md2wechat config show --format json
md2wechat config validate
md2wechat doctor --json
md2wechat config wechat-accounts --json
```

---

## 第一步：确认当前 CLI

先确认你运行的是当前安装的 CLI，而不是 PATH 上的旧版本：

```bash
which md2wechat
md2wechat version --json
md2wechat skills read md2wechat --json
```

`skills read` 读取的是当前二进制内置的 Agent SOP。它可以帮助 Agent 确认命令行为和本机 CLI 版本一致。

---

## 第二步：找到当前配置文件

查看当前生效配置：

```bash
md2wechat config show --format json
```

重点看这些字段：

```text
config_file
wechat_appid
wechat_account
md2wechat_api_key
md2wechat_base_url
default_convert_mode
image_provider
image_api_base
```

注意：`config show --format json` 输出的是扁平字段名；配置文件里是嵌套 YAML 键名。

| JSON 输出字段 | 配置文件字段 |
|---------------|--------------|
| `wechat_appid` | `wechat.appid` |
| `wechat_secret` | `wechat.secret` |
| `md2wechat_api_key` | `api.md2wechat_key` |
| `md2wechat_base_url` | `api.md2wechat_base_url` |
| `image_api_base` | `api.image_base_url` |

配置查找顺序：

1. `~/.config/md2wechat/config.yaml`
2. `~/.md2wechat.yaml`
3. `~/.md2wechat.yml`
4. 当前目录下的 `md2wechat.yaml` / `md2wechat.yml` / `md2wechat.json`
5. 当前目录下的 `.md2wechat.yaml` / `.md2wechat.yml` / `.md2wechat.json`

环境变量优先级高于配置文件。

---

## 第三步：初始化配置

首次使用：

```bash
md2wechat config init
```

默认生成：

```text
~/.config/md2wechat/config.yaml
```

如果你要为某个项目创建局部配置：

```bash
md2wechat config init ./md2wechat.yaml
```

不要把真实 `AppSecret`、`MD2WECHAT_API_KEY`、图片服务 key 提交到 Git。

---

## 第四步：单公众号配置

单公众号是默认主路径。配置文件写：

```yaml
wechat:
  appid: "你的微信公众号 AppID"
  secret: "你的微信公众号 AppSecret"

api:
  md2wechat_key: "你的 md2wechat API Key"
  md2wechat_base_url: "https://www.md2wechat.cn"
  convert_mode: "api"
  default_theme: "default"
```

也可以用环境变量临时覆盖微信凭证：

```bash
export WECHAT_APPID="你的微信公众号 AppID"
export WECHAT_SECRET="你的微信公众号 AppSecret"
```

验证：

```bash
md2wechat config validate
md2wechat config show --format json
md2wechat doctor --json
```

说明：

- `config validate` 只验证配置能加载和解析。
- `doctor` 是本地只读体检，不做 live auth、不上传图片、不创建草稿。
- 单公众号 direct 配置不会因为缺少 `MD2WECHAT_API_KEY` 而被多账号高级功能拦住。

---

## 第五步：多公众号配置

多公众号使用命名账号：

```yaml
wechat:
  default_account: main
  accounts:
    main:
      appid: "主公众号 AppID"
      secret: "主公众号 AppSecret"
    client-a:
      appid: "客户 A AppID"
      secret: "客户 A AppSecret"

api:
  md2wechat_key: "你的 md2wechat API Key"
  md2wechat_base_url: "https://www.md2wechat.cn"
```

账号名规则：

- 只能使用小写字母、数字、`_`、`-`
- 必须以小写字母或数字开头
- 示例：`main`、`client-a`、`brand_2026`

查看当前本地账号：

```bash
md2wechat config wechat-accounts --json
```

这个命令只读取本地配置：

- 不调用 `/api/auth/validate`
- 不要求 `MD2WECHAT_API_KEY`
- 不输出 `secret`

---

## 第六步：账号选择顺序

会调用微信接口的命令按这个顺序选择账号：

```text
--wechat-account
  -> WECHAT_ACCOUNT
  -> wechat.default_account
  -> 直接配置 wechat.appid / wechat.secret
  -> 唯一的命名账号
  -> WECHAT_ACCOUNT_AMBIGUOUS
```

显式选择账号：

```bash
md2wechat convert article.md --draft --cover cover.jpg --wechat-account client-a
md2wechat upload_image cover.jpg --wechat-account client-a
md2wechat create_image_post --title "标题" --images cover.jpg --wechat-account client-a
```

临时用环境变量选择：

```bash
export WECHAT_ACCOUNT=client-a
md2wechat convert article.md --draft --cover cover.jpg
```

如果你配置了多个命名账号，但没有 `default_account`，也没有传 `--wechat-account` 或 `WECHAT_ACCOUNT`，副作用命令会失败并返回 `WECHAT_ACCOUNT_AMBIGUOUS`。

---

## 第七步：API key 边界

`api.md2wechat_key` / `MD2WECHAT_API_KEY` 有两类用途：

1. API 模式转换 Markdown 到微信 HTML。
2. 命名账号执行微信副作用前，验证你已购买高级 API 服务。

命名账号执行这些副作用前会校验 API key：

- `upload_image`
- `download_and_upload`
- `generate_image`
- `generate_cover`
- `generate_infographic`
- `create_draft`
- `test-draft`
- `convert --upload`
- `convert --draft`
- `create_image_post` 非 dry-run

不会做 live API-key 校验的本地只读命令：

- `config show`
- `config validate`
- `doctor`
- `config wechat-accounts`
- `inspect`，除非它只是根据本地配置判断 readiness

命名账号 API key 校验请求：

```text
HEAD /api/auth/validate
Authorization: Bearer <MD2WECHAT_API_KEY>
```

如果你用 `api.md2wechat_base_url` 覆盖服务地址，校验也会使用同一个服务根地址。

---

## 第八步：执行前检查文章

创建草稿前，先检查单篇文章 readiness：

```bash
md2wechat inspect article.md --draft --json
```

Agent 应读取：

```text
data.readiness.targets
data.readiness.blockers
```

常见 blocker：

| blocker | 含义 |
|---------|------|
| `MISSING_API_KEY` | API 模式缺少 `MD2WECHAT_API_KEY` |
| `MISSING_COVER` | 草稿模式缺少 `--cover` 或 `--cover-media-id` |
| `LOCAL_IMAGE_MISSING` | 本地图片路径不存在 |

`doctor --json` 是本地配置体检；`inspect --json` 是单篇文章执行状态。不要混用。

---

## 第九步：常用命令

基础配置：

```bash
md2wechat config init
md2wechat config validate
md2wechat config show --format json
md2wechat doctor --json
```

多账号：

```bash
md2wechat config wechat-accounts --json
md2wechat convert article.md --draft --cover cover.jpg --wechat-account main
WECHAT_ACCOUNT=main md2wechat upload_image cover.jpg
```

转换和草稿：

```bash
md2wechat inspect article.md --draft --json
md2wechat convert article.md --preview
md2wechat convert article.md --output article.html
md2wechat convert article.md --draft --cover cover.jpg
```

图片：

```bash
md2wechat upload_image cover.jpg
md2wechat generate_image "封面图提示词"
```

---

## 第十步：常见错误

| 错误码 | 常见原因 | 处理方式 |
|--------|----------|----------|
| `WECHAT_ACCOUNT_NOT_FOUND` | 传了不存在的 `--wechat-account` 或 `WECHAT_ACCOUNT` | 运行 `md2wechat config wechat-accounts --json` 查看账号名 |
| `WECHAT_ACCOUNT_AMBIGUOUS` | 多个命名账号但没有默认选择 | 设置 `wechat.default_account`，或传 `--wechat-account` |
| `WECHAT_ACCOUNT_INVALID` | 账号名格式不合法 | 改成小写字母、数字、`_`、`-` |
| `API_KEY_REQUIRED` | 命名账号副作用缺少 `MD2WECHAT_API_KEY` | 配置 `api.md2wechat_key` 或导出 `MD2WECHAT_API_KEY` |
| `API_KEY_INVALID` | API key 被服务端判定无效 | 检查 key 是否复制完整，或联系服务提供方 |
| `API_KEY_VERIFY_FAILED` | 无法连接验证服务 | 检查网络、`api.md2wechat_base_url`、代理或服务状态 |
| `ip not in whitelist` | 微信后台没有放行当前机器公网 IP | 按 [微信凭证与 IP 白名单指南](WECHAT-CREDENTIALS.md) 配置白名单 |

---

## Agent 最小检查协议

当用户说“检查配置”“为什么不能发草稿”“帮我切账号”时，Agent 不要全量扫描所有 catalog。按任务运行最小命令。

配置排查：

```bash
md2wechat config show --format json
md2wechat config validate
md2wechat doctor --json
```

多账号排查：

```bash
md2wechat config wechat-accounts --json
```

单篇文章发布前：

```bash
md2wechat inspect article.md --draft --json
```

Agent 规则：

- 不输出 `secret`。
- 不替用户重置 `AppSecret`。
- 不自动运行会上传、创建草稿或修改远端状态的命令。
- 不把 `doctor` 的本地 readiness 当成单篇文章 readiness。
- 不把 `config validate` 当成 live auth 成功。
