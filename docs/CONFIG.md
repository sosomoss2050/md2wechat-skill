# 配置指南

这份文档解决 4 个最常见的问题：

1. 配置文件在哪里
2. 默认 API 域名在哪里改
3. Agent 应该先看哪里
4. 哪些功能分别需要哪些凭证

如果你只想先跑通主路径，先看下面这 3 步。

## 3 步完成基础配置

### 1. 生成示例配置

```bash
md2wechat config init
```

默认会生成到：

```text
~/.config/md2wechat/config.yaml
```

你也可以显式指定输出位置：

```bash
md2wechat config init ./md2wechat.yaml
```

### 2. 打开配置文件，先填最小必需项

```yaml
wechat:
  appid: "你的微信公众号 AppID"
  secret: "你的微信公众号 Secret"

api:
  md2wechat_key: "你的 md2wechat API Key"
  md2wechat_base_url: "https://www.md2wechat.cn"
  convert_mode: "api"
  default_theme: "default"
```

### 3. 验证当前配置

```bash
md2wechat config validate
md2wechat config show --format json
```

---

## Agent 和用户应该先看哪里

如果你不知道去哪改配置，按这个顺序找：

1. `~/.config/md2wechat/config.yaml`
2. 环境变量
3. 当前目录下的 `md2wechat.yaml` / `md2wechat.yml` / `md2wechat.json`

对 Agent 来说，**默认应该优先检查 `~/.config/md2wechat/config.yaml`**。
如果用户说“把 API 域名改成备用域名”“切换图片服务”“检查当前配置”，先运行：

```bash
md2wechat config show --format json
```

这样可以直接看到当前生效的：

- `config_file`
- `md2wechat_base_url`
- `image_provider`
- `image_api_base`
- `default_convert_mode`

---

## 默认 API 域名在哪里改

项目当前默认值是：

```text
https://www.md2wechat.cn
```

它**不是写死不可改**。你有两种常用改法。

### 方式一：改配置文件

编辑 `~/.config/md2wechat/config.yaml`：

```yaml
api:
  md2wechat_base_url: "https://www.md2wechat.cn"
```

如果你要切到备用域名：

```yaml
api:
  md2wechat_base_url: "https://md2wechat.app"
```

### 方式二：用环境变量临时覆盖

```bash
export MD2WECHAT_BASE_URL="https://md2wechat.app"
```

环境变量优先级高于配置文件，适合：

- 临时切换备用域名
- CI / Agent 自动化
- 不想修改全局配置文件的场景

---

## 配置文件搜索顺序

程序会按以下顺序查找配置文件：

1. `~/.config/md2wechat/config.yaml`
2. `~/.md2wechat.yaml`
3. `~/.md2wechat.yml`
4. `./md2wechat.yaml`
5. `./md2wechat.yml`
6. `./md2wechat.json`
7. `./.md2wechat.yaml`
8. `./.md2wechat.yml`
9. `./.md2wechat.json`

实践上建议：

- 全局默认配置放 `~/.config/md2wechat/config.yaml`
- 项目特殊配置再放当前目录

---

## 完整示例配置

仓库里提供了一份可直接参考的示例：

- [config.yaml.example](examples/config.yaml.example)

完整示例：

```yaml
wechat:
  appid: "your_wechat_appid"
  secret: "your_wechat_secret"

api:
  md2wechat_key: "your_md2wechat_api_key"
  md2wechat_base_url: "https://www.md2wechat.cn"
  image_key: "your_image_api_key"
  image_base_url: "https://api.openai.com/v1"
  image_provider: "openai"
  image_model: "dall-e-3"
  image_size: "1024x1024"
  convert_mode: "api"
  default_theme: "default"
  background_type: "default"
  http_timeout: 30

image:
  compress: true
  max_width: 1920
  max_size_mb: 5
```

---

## 配置项说明

### 微信配置

| 配置项 | 必需 | 说明 |
|--------|------|------|
| `wechat.appid` | 创建草稿、上传图片时需要 | 微信公众号 AppID |
| `wechat.secret` | 创建草稿、上传图片时需要 | 微信公众号 Secret |

### API 转换配置

| 配置项 | 必需 | 说明 | 默认值 |
|--------|------|------|--------|
| `api.md2wechat_key` | API 模式需要 | md2wechat API Key | - |
| `api.md2wechat_base_url` | 否 | 排版 API 域名 | `https://www.md2wechat.cn` |
| `api.convert_mode` | 否 | 默认转换模式 | `api` |
| `api.default_theme` | 否 | 默认主题 | `default` |
| `api.background_type` | 否 | 背景类型 | `default` |
| `api.http_timeout` | 否 | HTTP 超时秒数 | `30` |

### 图片生成配置

| 配置项 | 必需 | 说明 | 默认值 |
|--------|------|------|--------|
| `api.image_key` | AI 图片时需要 | 图片生成 API Key | - |
| `api.image_provider` | 否 | 图片服务提供方 | `openai` |
| `api.image_base_url` | 否 | 图片服务地址 | `https://api.openai.com/v1` |
| `api.image_model` | 否 | 图片模型 | `dall-e-3` |
| `api.image_size` | 否 | 默认图片尺寸 | `1024x1024` |

### 图片处理配置

| 配置项 | 必需 | 说明 | 默认值 |
|--------|------|------|--------|
| `image.compress` | 否 | 是否自动压缩 | `true` |
| `image.max_width` | 否 | 最大宽度 | `1920` |
| `image.max_size_mb` | 否 | 最大大小（MB） | `5` |

---

## 环境变量对照表

| 环境变量 | 对应配置项 |
|----------|------------|
| `WECHAT_APPID` | `wechat.appid` |
| `WECHAT_SECRET` | `wechat.secret` |
| `MD2WECHAT_API_KEY` | `api.md2wechat_key` |
| `MD2WECHAT_BASE_URL` | `api.md2wechat_base_url` |
| `IMAGE_API_KEY` | `api.image_key` |
| `IMAGE_API_BASE` | `api.image_base_url` |
| `IMAGE_PROVIDER` | `api.image_provider` |
| `IMAGE_MODEL` | `api.image_model` |
| `IMAGE_SIZE` | `api.image_size` |
| `CONVERT_MODE` | `api.convert_mode` |
| `DEFAULT_THEME` | `api.default_theme` |
| `DEFAULT_BACKGROUND_TYPE` | `api.background_type` |
| `HTTP_TIMEOUT` | `api.http_timeout` |
| `COMPRESS_IMAGES` | `image.compress` |
| `MAX_IMAGE_WIDTH` | `image.max_width` |
| `MAX_IMAGE_SIZE` | `image.max_size_mb` |

---

## 常见场景怎么配

### 只预览，不创建草稿

最小需要：

```yaml
api:
  md2wechat_key: "your_md2wechat_api_key"
  md2wechat_base_url: "https://www.md2wechat.cn"
  convert_mode: "api"
```

### 需要上传图片和创建草稿

最小需要：

```yaml
wechat:
  appid: "your_wechat_appid"
  secret: "your_wechat_secret"

api:
  md2wechat_key: "your_md2wechat_api_key"
```

### 需要 AI 图片生成

最小需要：

```yaml
wechat:
  appid: "your_wechat_appid"
  secret: "your_wechat_secret"

api:
  image_key: "your_image_api_key"
  image_provider: "modelscope"
  image_base_url: "https://api-inference.modelscope.cn"
```

---

## 配置优先级

优先级从高到低：

```text
命令行参数 > 环境变量 > 配置文件 > 默认值
```

举例：

1. 配置文件里写了：

```yaml
api:
  md2wechat_base_url: "https://www.md2wechat.cn"
```

2. 当前终端又执行了：

```bash
export MD2WECHAT_BASE_URL="https://md2wechat.app"
```

最终生效的是：

```text
https://md2wechat.app
```

---

## 自检命令

```bash
md2wechat config init
md2wechat config show --format json
md2wechat config validate
```

推荐排查顺序：

1. 先看 `config_file` 指向哪个文件
2. 再看 `md2wechat_base_url` 是否真是你想要的域名
3. 再看 `image_provider` / `image_api_base` 是否匹配
4. 最后检查环境变量是否把文件里的值覆盖掉了

---

## 相关文档

- [新手快速开始](QUICKSTART.md)
- [安装指南](INSTALL.md)
- [图片服务配置](IMAGE_PROVISIONERS.md)
- [真实烟雾测试记录](SMOKE.md)
