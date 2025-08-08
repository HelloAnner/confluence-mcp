# Confluence MCP Server (Go版本)

这是一个使用 Go 语言和 mcp-go 框架实现的 Confluence MCP (Model Context Protocol) 服务器，可以让 AI 助手直接与 Confluence 进行交互。

## 功能特性

- 🔍 **获取页面内容** - 根据页面ID获取完整的页面信息和评论
- 📄 **获取子页面** - 获取指定页面的所有子页面列表
- ✏️ **创建页面** - 在指定空间创建新的 Confluence 页面
- 💬 **创建评论** - 为页面添加评论
- 🔎 **搜索页面** - 根据关键字搜索 Confluence 页面

## 环境要求

- Go 1.21 或更高版本
- Confluence Cloud 或 Server 实例
- Confluence API Token

## 安装和配置

### 1. 克隆项目

```bash
git clone <repository-url>
cd confluence-mcp
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置环境变量

创建 `.env` 文件并配置以下环境变量：

```env
CONFLUENCE_BASE_URL=https://your-domain.atlassian.net
CONFLUENCE_API_TOKEN=your-api-token
CONFLUENCE_USER_EMAIL=your-email@example.com
```

### 4. 获取 Confluence API Token

1. 登录到 [Atlassian Account Settings](https://id.atlassian.com/manage-profile/security/api-tokens)
2. 点击 "Create API token"
3. 输入标签名称并创建
4. 复制生成的 token 到 `.env` 文件中

## 使用方法

### 直接运行

```bash
go run .
```

### 编译后运行

```bash
go build -o confluence-mcp
./confluence-mcp
```

### 在 Cursor 中使用

#### 本地调用

如果已在本地编译安装了 confluence-mcp，在 Cursor 的 MCP 配置中添加：

```json
{
  "mcpServers": {
    "confluence": {
      "command": "/usr/local/bin/confluence-mcp",
      "args": [],
      "env": {
        "CONFLUENCE_BASE_URL": "",
        "CONFLUENCE_API_TOKEN": "",
        "CONFLUENCE_USER_EMAIL": ""
      }
    }
  }
}
```

#### 局域网连接

如果想通过局域网连接到运行 confluence-mcp 的服务器，可以使用以下配置：

```json
{
  "mcpServers": {
    "confluence": {
      "url": "http://<服务器IP>:8080",
      "env": {
        "CONFLUENCE_BASE_URL": "",
        "CONFLUENCE_API_TOKEN": "",
        "CONFLUENCE_USER_EMAIL": ""
      }
    }
  }
}
```

#### 远程连接

如果想通过远程连接到运行 confluence-mcp 的服务器，可以使用以下配置：

```json
{
  "mcpServers": {
    "remote-confluence": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

注意事项：
1. 将 `<服务器IP>` 替换为运行 confluence-mcp 的服务器 IP 地址
2. 确保服务器防火墙允许 8080 端口访问
3. 服务器端需要使用以下命令启动服务：
   ```bash
   ./confluence-mcp --host 0.0.0.0 --port 8080
   ```

## 可用工具

### 1. get_page
获取页面内容和评论

**参数：**
- `pageId` (必需): Confluence 页面 ID
- `includeComments` (可选): 是否包含评论，默认为 true

### 2. get_child_pages
获取子页面列表

**参数：**
- `pageId` (必需): 父页面 ID
- `limit` (可选): 返回结果数量限制，默认为 50

### 3. create_page
创建新页面

**参数：**
- `spaceKey` (必需): 空间键值
- `title` (必需): 页面标题
- `content` (必需): 页面内容（Confluence 存储格式）
- `parentId` (可选): 父页面 ID

### 4. create_comment
创建页面评论

**参数：**
- `pageId` (必需): 页面 ID
- `comment` (必需): 评论内容

### 5. search_pages
搜索页面

**参数：**
- `query` (必需): 搜索关键字
- `spaceKey` (可选): 限制搜索的空间
- `limit` (可选): 返回结果数量限制，默认为 20

## 项目结构

```
confluence-mcp/
├── main.go          # 主程序入口和工具注册
├── confluence.go    # Confluence API 客户端
├── handlers.go      # MCP 工具处理函数
├── go.mod          # Go 模块定义
├── .env            # 环境变量配置
└── README.md       # 项目说明
```

## 技术栈

- **Go 1.21+** - 主要编程语言
- **mcp-go** - MCP 协议实现框架
- **godotenv** - 环境变量管理
- **Confluence REST API** - 与 Confluence 交互

## 错误处理

服务器包含完善的错误处理机制：
- API 请求失败时返回详细错误信息
- 参数验证确保请求的有效性
- 网络超时和重试机制

## 安全注意事项

- 请妥善保管 API Token，不要提交到版本控制系统
- 建议使用具有适当权限的专用账户
- 在生产环境中使用 HTTPS

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！