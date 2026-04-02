# Key Agent 用户指南

本指南涵盖使用 Key Agent 管理键值数据和密钥的所有方面。

## 目录

- [概述](#概述)
- [安装](#安装)
- [配置](#配置)
- [CLI 命令参考](#cli-命令参考)
- [HTTP API](#http-api)
- [Go SDK](#go-sdk)
- [MCP 集成](#mcp-集成)
- [安全最佳实践](#安全最佳实践)
- [故障排除](#故障排除)

## 概述

Key Agent 是一个本地守护进程，提供键值数据和密钥的安全存储。主要特性：

- **加密存储** 使用 AES-256-GCM
- **基于令牌的认证** 用于 API 访问
- **CLI 工具** (`keyctl`) 便于管理
- **Go SDK** 用于编程访问
- **MCP 支持** 用于 AI 代理集成

## 安装

### Docker（推荐）

使用 Docker 是运行 Key Agent 最简单的方式：

```bash
# 克隆仓库
git clone https://github.com/skys-mission/key-agent.git
cd key-agent

# 使用 Docker Compose 启动
docker compose up -d

# 检查服务状态
docker compose ps

# 查看日志（首次运行包含 root 令牌）
docker compose logs key-agent

# 停止服务
docker compose down
```

**Docker 配置说明：**

`docker-compose.yml` 文件包含：
- 数据卷持久化
- 健康检查
- 基于文件的主密钥后端（容器中必需）

**环境变量：**

| 变量 | 说明 |
|------|------|
| `KEY_AGENT_MASTER_KEY_BACKEND` | 容器中设置为 `file` |
| `KEY_AGENT_PASSPHRASE` | 文件后端的密码短语（推荐设置） |

**手动 Docker 运行：**

```bash
# 构建镜像
docker build -t key-agent:latest .

# 运行容器
docker run -d \
  --name key-agent \
  -p 127.0.0.1:8080:8080 \
  -v key-agent-data:/data \
  -e KEY_AGENT_MASTER_KEY_BACKEND=file \
  -e KEY_AGENT_PASSPHRASE=your-secure-passphrase \
  key-agent:latest
```

### 二进制下载

从 [GitHub Releases](https://github.com/skys-mission/key-agent/releases) 下载最新版本：

```bash
# 下载守护进程
curl -sL https://github.com/skys-mission/key-agent/releases/latest/download/key-agent-linux-amd64 -o key-agent
chmod +x key-agent
sudo mv key-agent /usr/local/bin/

# 下载 CLI
curl -sL https://github.com/skys-mission/key-agent/releases/latest/download/keyctl-linux-amd64 -o keyctl
chmod +x keyctl
sudo mv keyctl /usr/local/bin/
```

### 从源码构建

```bash
git clone https://github.com/skys-mission/key-agent.git
cd key-agent
make build
```

## 配置

### 默认位置

| 文件 | 位置 |
|------|------|
| 守护进程配置 | `~/.skys-mission/key-agent/config.yaml` |
| CLI 配置 | `~/.skys-mission/key-agent/keyctl.yaml` |
| 数据目录 | `~/.skys-mission/key-agent/data/` |
| 令牌文件 | `~/.skys-mission/key-agent/data/token` |

### 配置文件

创建 `~/.skys-mission/key-agent/config.yaml`：

```yaml
# 服务设置
server:
  addr: "127.0.0.1:8080"

# 存储设置
storage:
  data_dir: "~/.skys-mission/key-agent/data"
  db_name: "key-agent.db"

# 安全设置
security:
  master_key_backend: "auto"  # auto, keyring, tpm, file

# 日志设置
logging:
  level: "info"           # debug, info, warn, error
  format: "json"          # json 或 text
  file: ""                # 空表示 stderr，或日志文件路径
  max_size: 100           # 轮转前最大大小（MB）
  max_backups: 3          # 旧日志文件最大数量
  max_age: 30             # 保留旧日志的最大天数
  compress: true          # 压缩轮转的日志

# MCP 设置
mcp:
  enabled: true
  endpoint: "/mcp"
```

### 主密钥后端选项

| 后端 | 描述 | 平台 |
|------|------|------|
| `auto` | 自动选择最佳可用选项 | 所有 |
| `keyring` | 系统密钥环（Keychain/Secret Service） | macOS, Linux |
| `tpm` | 可信平台模块 | Linux |
| `file` | 加密文件（安全性较低） | 所有 |

## CLI 配置

`keyctl` CLI 支持持久化配置，避免每次命令都输入 `--addr` 和 `--token`。

### 配置文件位置

文件路径：`~/.skys-mission/key-agent/keyctl.yaml`

### 配置命令

```bash
# 设置默认服务器地址
keyctl config set addr http://127.0.0.1:8080

# 设置默认令牌
keyctl config set token ka_xxxxxxxx...

# 设置令牌文件路径（作为内联令牌的替代）
keyctl config set token_file ~/.mytoken

# 获取配置值
keyctl config get addr

# 列出所有配置
keyctl config list
```

### 配置优先级

配置按以下优先级解析（从高到低）：

1. **命令行参数**：`--addr`、`--token`、`--token-file`
2. **环境变量**：`KEY_AGENT_ADDR`、`KEY_AGENT_TOKEN`
3. **配置文件**：`~/.skys-mission/key-agent/keyctl.yaml`

### 环境变量

| 变量 | 说明 |
|------|------|
| `KEY_AGENT_ADDR` | 服务器地址（如 `http://127.0.0.1:8080`） |
| `KEY_AGENT_TOKEN` | API 认证令牌 |

### 示例工作流

```bash
# 1. 启动守护进程
key-agent

# 2. 保存输出中的 root 令牌
# Root token: ka_abc123...

# 3. 一次性配置 CLI
keyctl config set addr http://127.0.0.1:8080
keyctl config set token ka_abc123...

# 4. 无需参数即可使用 CLI
keyctl kv set mykey myvalue
keyctl kv get mykey
keyctl secret set db/password secret123 --type password
```

## CLI 命令参考

### 启动守护进程

```bash
# 使用默认配置启动
key-agent

# 使用自定义配置启动
key-agent --config /path/to/config.yaml

# 使用自定义数据目录启动
key-agent --data-dir /custom/data/path
```

### KV 操作

```bash
# 设置值
keyctl kv set <key> <value>

# 示例
keyctl kv set app/database/host "localhost"
keyctl kv set app/database/port "5432"
keyctl kv set app/feature/flags '{"enabled": true}'

# 获取值
keyctl kv get <key>

# 获取原始值（无 JSON 格式化）
keyctl kv get <key> --raw

# 列出所有键
keyctl kv list

# 按前缀过滤
keyctl kv list --prefix app/database/

# 删除键
keyctl kv delete <key>
```

### 密钥操作

```bash
# 设置密钥
keyctl secret set <key> <value> --type <type>

# 密钥类型: password, api_key, certificate, private_key, token, other

# 示例
keyctl secret set db/password "mysecretpass" --type password
keyctl secret set openai/api_key "sk-xxx" --type api_key
keyctl secret set aws/access_key "AKIAxxx" --type api_key

# 获取密钥
keyctl secret get <key>

# 列出密钥
keyctl secret list

# 按前缀列出
keyctl secret list --prefix db/

# 删除密钥
keyctl secret delete <key>
```

### 令牌操作

```bash
# 保存初始令牌
keyctl token save <token>

# 创建新令牌
keyctl token create --name <name> [--type <type>] [--expires-in <duration>]

# 示例
keyctl token create --name "my-app" --type client --expires-in 24h
keyctl token create --name "mcp-agent" --type mcp --expires-in 30d
```

### CLI 选项

```bash
# 指定服务器地址
keyctl --addr http://localhost:8080 kv get mykey

# 直接指定令牌
keyctl --token "ka_xxx" kv get mykey

# 指定令牌文件
keyctl --token-file /path/to/token kv get mykey
```

## HTTP API

### 基础 URL

```
http://127.0.0.1:8080
```

所有 API 请求需要 Bearer 令牌认证：

```
Authorization: Bearer <token>
```

### 端点

#### 健康检查

```
GET /health
```

响应：
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

#### KV 操作

```
GET    /api/v1/kv/<key>     # 获取值
PUT    /api/v1/kv/<key>     # 设置值
DELETE /api/v1/kv/<key>     # 删除键
GET    /api/v1/kv           # 列出键 (使用 ?prefix=...)
```

设置 KV 请求：
```json
{
  "value": "my-value",
  "metadata": {
    "description": "可选描述"
  }
}
```

#### 密钥操作

```
GET    /api/v1/secrets/<key>     # 获取密钥
PUT    /api/v1/secrets/<key>     # 设置密钥
DELETE /api/v1/secrets/<key>     # 删除密钥
GET    /api/v1/secrets           # 列出密钥 (使用 ?prefix=...)
```

设置密钥请求：
```json
{
  "value": "secret-value",
  "type": "password",
  "metadata": {
    "description": "数据库密码"
  }
}
```

#### 令牌操作

```
POST /api/v1/token     # 创建新令牌
```

创建令牌请求：
```json
{
  "name": "my-app",
  "type": "client",
  "expires_in": "24h"
}
```

## Go SDK

### 安装

```bash
go get github.com/skys-mission/keysdk
```

### 使用

```go
package main

import (
    "fmt"

    "github.com/skys-mission/keysdk"
)

func main() {
    // 创建客户端
    client := keysdk.NewClient(&keysdk.Config{
        BaseURL: "http://127.0.0.1:8080",
        Token:   "your-token-here",
    })

    // KV 示例
    kvExample(client)

    // 密钥示例
    secretExample(client)

    // 令牌示例
    tokenExample(client)
}

func kvExample(client *keysdk.Client) {
    // 设置
    entry, err := client.SetKV("my-key", &keysdk.SetKVOptions{
        Value: "my-value",
        Metadata: map[string]interface{}{
            "description": "示例键",
        },
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("设置: %+v\n", entry)

    // 获取
    entry, err = client.GetKV("my-key")
    if err != nil {
        panic(err)
    }
    fmt.Printf("获取: %+v\n", entry)

    // 列出
    keys, err := client.ListKV("my-")
    if err != nil {
        panic(err)
    }
    fmt.Printf("键: %v\n", keys)

    // 删除
    client.DeleteKV("my-key")
}

func secretExample(client *keysdk.Client) {
    // 设置密钥
    secret, err := client.SetSecret("db-password", &keysdk.SetSecretOptions{
        Value: "super-secret",
        Type:  keysdk.SecretTypePassword,
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("密钥: %+v\n", secret)

    // 获取密钥
    secret, _ = client.GetSecret("db-password")
    fmt.Printf("密钥值: %s\n", secret.Value)
}

func tokenExample(client *keysdk.Client) {
    // 创建令牌
    token, err := client.CreateToken(&keysdk.CreateTokenOptions{
        Name:      "my-app",
        Type:      "client",
        ExpiresIn: "24h",
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("新令牌: %s\n", token.Token)
}
```

## MCP 集成

Key Agent 支持模型上下文协议 (MCP) 用于 AI 代理集成。

### 启用 MCP

在配置中：

```yaml
mcp:
  enabled: true
  endpoint: /mcp
```

### 配置 AI 助手

对于 Claude Desktop 或其他 MCP 客户端，添加到配置中：

```json
{
  "mcpServers": {
    "key-agent": {
      "url": "http://127.0.0.1:8080/mcp",
      "headers": {
        "Authorization": "Bearer your-token-here"
      }
    }
  }
}
```

### 可用 MCP 工具

| 工具 | 描述 |
|------|------|
| `kv_get` | 获取 KV 值 |
| `kv_set` | 设置 KV 值 |
| `kv_delete` | 删除 KV 条目 |
| `kv_list` | 列出 KV 键 |
| `secret_get` | 获取密钥 |
| `secret_set` | 设置密钥 |
| `secret_delete` | 删除密钥 |
| `secret_list` | 列出密钥键 |

## 安全最佳实践

### 令牌管理

1. **安全保存 root 令牌** - 它只会显示一次
2. **创建受限令牌** - 为临时访问设置过期时间
3. **定期轮换令牌** - 删除旧令牌并创建新令牌

### 网络安全

1. **默认绑定** - 仅绑定到 localhost (127.0.0.1)
2. **无 TLS** - 设计用于本地使用；远程访问请使用反向代理
3. **防火墙规则** - 如果绑定到外部接口，请限制访问

### 存储安全

1. **主密钥** - 可用时使用 `keyring` 后端
2. **备份** - 安全备份数据目录
3. **文件权限** - 确保 `~/.skys-mission/key-agent/` 具有限制性权限 (700)

## 故障排除

### 守护进程无法启动

```bash
# 检查端口是否被占用
lsof -i :8080

# 检查日志
key-agent --log-level debug
```

### 令牌不工作

```bash
# 验证令牌已保存
cat ~/.skys-mission/key-agent/data/token

# 尝试显式指定令牌
keyctl --token "ka_xxx" kv list
```

### 权限被拒绝

```bash
# 修复权限
chmod 700 ~/.skys-mission/key-agent
chmod 600 ~/.skys-mission/key-agent/data/token
```

### Linux 上密钥环问题

```bash
# 安装 dbus 和 secret-service
sudo apt install dbus libsecret-1-0

# 或使用文件后端
key-agent --master-key-backend file
```

## 获取帮助

- **GitHub Issues**: [github.com/skys-mission/key-agent/issues](https://github.com/skys-mission/key-agent/issues)
- **文档**: [docs/](https://github.com/skys-mission/key-agent/tree/main/docs)
