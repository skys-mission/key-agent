<p align="center">
  <img src="https://img.shields.io/github/v/release/skys-mission/key-agent?include_prereleases" alt="Release">
  <img src="https://img.shields.io/github/go-mod/go-version/skys-mission/key-agent" alt="Go Version">
  <img src="https://img.shields.io/github/license/skys-mission/key-agent" alt="License">
  <img src="https://img.shields.io/github/actions/workflow/status/skys-mission/key-agent/ci.yml?branch=main" alt="CI Status">
  <img src="https://codecov.io/gh/skys-mission/key-agent/branch/main/graph/badge.svg" alt="Coverage">
</p>

<h1 align="center">🔑 Key Agent</h1>

<p align="center">
  <strong>轻量级、安全的本地键值对和密钥管理守护进程</strong><br>
  <sub>面向开发者、运维人员和 AI 代理</sub>
</p>

<p align="center">
  <a href="#-特性">特性</a> •
  <a href="#-快速开始">快速开始</a> •
  <a href="#-安装">安装</a> •
  <a href="#-使用方法">使用方法</a> •
  <a href="#-sdk">SDK</a> •
  <a href="../README.md">English</a>
</p>

---

## ✨ 特性

- 🔐 **加密存储** - AES-256-GCM 加密，主密钥存储在系统密钥环、TPM 或文件中
- 🤖 **MCP 支持** - 通过 [模型上下文协议](https://modelcontextprotocol.io/) 原生集成 AI 代理
- 🖥️ **CLI 工具** - 简单直观的命令行工具 (`keyctl`)
- 🔌 **Go SDK** - 从 Go 应用程序编程访问
- 📝 **结构化日志** - JSON 格式日志，支持轮转和大小限制
- 🚀 **单二进制文件** - 零外部依赖，易于部署
- 🔄 **令牌管理** - 创建和管理带过期时间的访问令牌

## 📦 安装

### Docker（推荐快速开始）

```bash
# 1. 克隆并启动
git clone https://github.com/skys-mission/key-agent.git
cd key-agent
docker compose up -d

# 2. 从日志获取 root 令牌
docker compose logs key-agent | grep "Root token"
# 输出: Root token: ka_xxxxxxxx...

# 3. 配置 CLI（一次性设置）
keyctl config set addr http://127.0.0.1:8080
keyctl config set token ka_xxxxxxxx...

# 4. 测试
keyctl kv set hello world
keyctl kv get hello
```

**Docker 配置：**

| 变量 | 描述 |
|------|------|
| `KEY_AGENT_MASTER_KEY` | Base64 编码的 32 字节主密钥 |
| `KEY_AGENT_PASSPHRASE` | 加密主密钥文件的密码短语 |

数据通过卷挂载存储在宿主机的 `~/.skys-mission/key-agent/data/` 目录。

### 从发布版本安装

```bash
# macOS / Linux
curl -sL https://github.com/skys-mission/key-agent/releases/latest/download/key-agent-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m) -o key-agent
chmod +x key-agent
sudo mv key-agent /usr/local/bin/

# CLI 工具
curl -sL https://github.com/skys-mission/key-agent/releases/latest/download/keyctl-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m) -o keyctl
chmod +x keyctl
sudo mv keyctl /usr/local/bin/
```

### 从源码构建

```bash
git clone https://github.com/skys-mission/key-agent.git
cd key-agent
make build
sudo make install
```

## 🚀 快速开始

### 方式一：Docker（最快）

```bash
# 克隆并启动
git clone https://github.com/skys-mission/key-agent.git
cd key-agent
docker compose up -d

# 检查服务健康状态
curl http://127.0.0.1:8080/health

# 查看日志获取 root 令牌
docker logs key-agent
```

### 方式二：二进制文件

#### 步骤 1：启动守护进程

```bash
key-agent
```

首次运行时会生成并显示 root 令牌。**请安全保存此令牌！**

```
========================================
Root token generated (save this token):
ka_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
========================================
```

#### 步骤 2：配置 CLI

```bash
# 一次性配置
keyctl config set addr http://127.0.0.1:8080
keyctl config set token ka_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# 验证配置
keyctl config list
```

#### 步骤 3：存储和获取值

```bash
# 存储键值对
keyctl kv set app/database/host "localhost"
keyctl kv set app/database/port "5432"

# 获取值
keyctl kv get app/database/host
# 输出: localhost

# 列出所有键
keyctl kv list
```

#### 步骤 4：存储密钥

```bash
# 存储 API 密钥
keyctl secret set openai/api_key "sk-xxxxxxxx" --type api_key

# 存储密码
keyctl secret set db/postgres/password "mysecretpass" --type password

# 获取密钥
keyctl secret get openai/api_key
```

## 📖 使用方法

### 键值操作

```bash
# 设置值
keyctl kv set config/timeout "30s"

# 获取原始值
keyctl kv get config/timeout --raw

# 按前缀列出
keyctl kv list --prefix app/

# 删除键
keyctl kv delete config/timeout
```

### 密钥操作

```bash
# 可用类型: password, api_key, certificate, private_key, token, other
keyctl secret set aws/access_key "AKIAxxxx" --type api_key
keyctl secret set ssh/private_key "-----BEGIN..." --type private_key

# 列出密钥
keyctl secret list
```

### 令牌管理

```bash
# 创建新令牌
keyctl token create --name "my-app" --type client --expires-in 24h

# 保存令牌到文件
keyctl token save ka_xxxx
```

## 🔌 SDK

Key Agent 提供 Go SDK 用于编程访问：

```go
package main

import (
    "fmt"
    "github.com/skys-mission/keysdk"
)

func main() {
    client := keysdk.NewClient(&keysdk.Config{
        BaseURL: "http://127.0.0.1:8080",
        Token:   "your-token-here",
    })

    // 设置值
    entry, _ := client.SetKV("my-key", &keysdk.SetKVOptions{
        Value: "my-value",
    })
    fmt.Printf("已创建: %s\n", entry.Key)

    // 获取值
    entry, _ = client.GetKV("my-key")
    fmt.Printf("值: %s\n", entry.Value)

    // 存储密钥
    secret, _ := client.SetSecret("db-password", &keysdk.SetSecretOptions{
        Value: "super-secret",
        Type:  keysdk.SecretTypePassword,
    })
    fmt.Printf("密钥类型: %s\n", secret.Type)
}
```

## 🤖 MCP 集成

Key Agent 支持 [模型上下文协议 (MCP)](https://modelcontextprotocol.io/) 用于 AI 代理集成。在配置中启用：

```yaml
mcp:
  enabled: true
  endpoint: /mcp
```

配置你的 AI 助手连接到 `http://localhost:8080/mcp` 并使用有效的令牌。

## ⚙️ 配置

### 守护进程配置

文件：`~/.skys-mission/key-agent/config.yaml`

```yaml
server:
  addr: 127.0.0.1:8080

storage:
  data_dir: ~/.skys-mission/key-agent/data
  db_name: key-agent.db

security:
  master_key_backend: auto  # auto, keyring, tpm, file

logging:
  level: info
  format: json
  file: ""
  max_size: 100       # MB
  max_backups: 3
  max_age: 30         # 天
  compress: true

mcp:
  enabled: true
  endpoint: /mcp
```

### CLI 配置

文件：`~/.skys-mission/key-agent/keyctl.yaml`

一次性配置，随处使用：

```bash
# 设置默认服务器地址
keyctl config set addr http://127.0.0.1:8080

# 设置默认令牌
keyctl config set token ka_xxxxxxxx...

# 或设置令牌文件路径
keyctl config set token_file ~/.mytoken

# 查看当前配置
keyctl config list
```

**配置优先级**（从高到低）：
1. 命令行参数：`--addr`、`--token`
2. 环境变量：`KEY_AGENT_ADDR`、`KEY_AGENT_TOKEN`
3. 配置文件：`~/.skys-mission/key-agent/keyctl.yaml`

**环境变量：**

| 变量 | 说明 |
|------|------|
| `KEY_AGENT_ADDR` | 服务器地址（如 http://127.0.0.1:8080） |
| `KEY_AGENT_TOKEN` | API 认证令牌 |
| `KEY_AGENT_PASSPHRASE` | 文件主密钥密码短语 |

## 🛡️ 安全性

- **加密**：所有数据使用 AES-256-GCM 静态加密
- **主密钥**：安全存储在系统密钥环（macOS Keychain、Linux Secret Service）
- **令牌认证**：所有 API 操作需要有效令牌
- **无网络暴露**：默认仅绑定到 localhost

## 📚 文档

| 文档 | 描述 |
|------|------|
| [用户指南](USER_GUIDE_CN.md) | 详细使用说明 |
| [API 参考](API.md) | HTTP API 文档 |
| [开发指南](DEVELOPMENT.md) | 贡献和开发 |

## 🤝 贡献

欢迎贡献！请查看 [贡献指南](../CONTRIBUTING.md)。

```bash
# 运行测试
make test

# 运行代码检查
make lint

# 构建
make build
```

## 📄 许可证

[MIT 许可证](../LICENSE) 

---

<p align="center">
  95% Vibe Coding ✨ 