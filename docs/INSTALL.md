# 安装指南

本文档详细说明 md2wechat 的各种安装方式。

## 目录

- [系统要求](#系统要求)
- [方式一：安装脚本](#方式一安装脚本推荐)
- [方式二：预编译二进制](#方式二预编译二进制)
- [方式三：Go 工具链安装](#方式三go-工具链安装适合开发者)
- [方式四：从源码编译](#方式四从源码编译)
- [验证安装](#验证安装)
- [卸载](#卸载)

---

## 系统要求

- **操作系统**：Linux / macOS / Windows
- **Go 版本**：1.26.1 或更高（如果从源码编译）
- **网络**：需要访问微信公众号 API

---

## 方式一：安装脚本（推荐，固定版本 release 资产）

### macOS / Linux

```bash
export MD2WECHAT_RELEASE_BASE_URL=https://github.com/geekjourneyx/md2wechat-skill/releases/download/v2.0.1
curl -fsSL "${MD2WECHAT_RELEASE_BASE_URL}/install.sh" | bash
```

### Windows PowerShell

```powershell
$env:MD2WECHAT_RELEASE_BASE_URL = "https://github.com/geekjourneyx/md2wechat-skill/releases/download/v2.0.1"
iex ((New-Object System.Net.WebClient).DownloadString("$env:MD2WECHAT_RELEASE_BASE_URL/install.ps1"))
```

安装脚本会下载对应 release 资产并校验 `checksums.txt`。主路径应始终指向固定版本的 release，不要使用 `latest` 或 `main` 作为安装入口。

---

## 方式二：预编译二进制

### 下载地址

访问 [Releases](https://github.com/geekjourneyx/md2wechat-skill/releases) 页面下载与你目标版本匹配的资产。

| 系统 | 文件名 |
|------|--------|
| Linux (amd64) | `md2wechat-linux-amd64` |
| Linux (arm64) | `md2wechat-linux-arm64` |
| macOS (Intel) | `md2wechat-darwin-amd64` |
| macOS (Apple Silicon) | `md2wechat-darwin-arm64` |
| Windows (64位) | `md2wechat-windows-amd64.exe` |
| Installer shell script | `install.sh` |
| Installer PowerShell script | `install.ps1` |

### 安装步骤

#### Linux / macOS

```bash
VERSION=v2.0.1
ASSET=md2wechat-linux-amd64
# macOS 请改成 md2wechat-darwin-amd64 或 md2wechat-darwin-arm64
curl -LO https://github.com/geekjourneyx/md2wechat-skill/releases/download/${VERSION}/${ASSET}
curl -LO https://github.com/geekjourneyx/md2wechat-skill/releases/download/${VERSION}/checksums.txt
sha256sum -c checksums.txt --ignore-missing

# 2. 添加执行权限
chmod +x "${ASSET}"

# 3. 移动到 PATH
sudo mv "${ASSET}" /usr/local/bin/md2wechat

# 4. 验证
md2wechat version --json
```

#### Windows

```powershell
# 1. 下载与你目标版本匹配的 release 资产
# 2. 同时下载 install.ps1 / install.sh 和 checksums.txt 并校验

# 3. 验证
md2wechat.exe version --json
```

---

## 方式三：Go 工具链安装（适合开发者）

如果你正在本地开发、调试或需要最新源码，可以使用 `go install`：

```bash
go install github.com/geekjourneyx/md2wechat-skill/cmd/md2wechat@v2.0.1
```

如果你只是要稳定使用，仍然建议优先使用上面的安装脚本或 release 资产。

---

## 方式四：从源码编译

```bash
git clone https://github.com/geekjourneyx/md2wechat-skill.git
cd md2wechat-skill
go build -o md2wechat ./cmd/md2wechat
```

如果你需要交叉编译，可以直接设置 `GOOS` / `GOARCH` 后再运行 `go build`。

---

## 关于 Docker

仓库当前不提供官方 Docker 镜像或 Dockerfile。若后续补齐容器化发布流程，再以 release 文档为准。

---

## 验证安装

运行以下命令验证安装成功：

```bash
# 查看帮助
md2wechat version --json
md2wechat --help

# 查看所有命令
md2wechat help

# 配置校验是独立动作，不是安装验证的一部分
```

预期输出：

```
md2wechat converts Markdown articles to WeChat Official Account format
...
```

---

## 卸载

### Go 工具链安装

```bash
rm $(go env GOPATH)/bin/md2wechat
```

### 预编译二进制

```bash
# Linux/macOS
sudo rm /usr/local/bin/md2wechat

# Windows
# 删除安装目录下的 md2wechat.exe
```

---

## 下一步

安装完成后，请继续阅读 [配置指南](CONFIG.md) 和 [使用教程](USAGE.md)。
