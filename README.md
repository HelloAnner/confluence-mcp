# Confluence MCP Server 使用示例

## 前置准备

### 1. 配置环境变量

创建 `.env` 文件：

```bash
cp .env.example .env
```

编辑 `.env` 文件，填入你的 Confluence 配置：

```env
CONFLUENCE_BASE_URL=https://your-company.atlassian.net/wiki
CONFLUENCE_API_TOKEN=ATATT3xFfGF0T4JVjdmA1b2c3d4e5f6g7h8i9j0k
CONFLUENCE_USER_EMAIL=your.email@company.com
```

### 2. 启动服务器

#### 标准 MCP 模式（stdio）

```bash
# 开发模式
npm run mcp

# 或者构建后运行
npm run build
npm run start:mcp
```

#### SSE HTTP 服务模式（局域网访问）

```bash
# 开发模式
npm run server

# 或者构建后运行
npm run build
npm start
```

服务启动后，可通过以下端点访问：
- **健康检查**: http://localhost:8080/health
- **SSE 连接**: http://localhost:8080/sse
- **工具列表**: http://localhost:8080/mcp/tools
- **工具调用**: http://localhost:8080/mcp/call

#### Docker 部署（推荐用于生产环境）

```bash
# 使用 Docker Compose
docker compose up -d

# 或者直接使用 Docker
docker build -t confluence-mcp .
docker run -d -p 8080:8080 --env-file .env confluence-mcp
```

详细部署说明请参考 [DEPLOYMENT.md](./DEPLOYMENT.md)。

## 连接方式

### 1. 标准 MCP 客户端集成（stdio）

#### 使用 Claude Desktop 集成

**本地标准配置（stdio 模式）：**

```json
{
  "mcpServers": {
    "confluence": {
      "command": "node",
      "args": ["/path/to/confluence-mcp/dist/main.js"],
      "env": {
        "CONFLUENCE_BASE_URL": "https://your-company.atlassian.net/wiki",
        "CONFLUENCE_API_TOKEN": "your_api_token",
        "CONFLUENCE_USER_EMAIL": "your.email@company.com"
      }
    }
  }
}
```

**局域网 HTTP 配置（SSE 模式）：**

如果你的 Confluence MCP 服务器运行在局域网的另一台机器上，可以使用 HTTP 连接方式：

```json
{
  "mcpServers": {
    "confluence": {
      "command": "npx",
      "args": ["@modelcontextprotocol/server-fetch"],
      "env": {
        "FETCH_BASE_URL": "http://192.168.1.100:8080/mcp"
      }
    }
  }
}
```

或者使用 MCP HTTP 客户端：

```json
{
  "mcpServers": {
    "confluence": {
      "command": "node",
      "args": ["-e", "require('@modelcontextprotocol/sdk/client/http').createHttpClient('http://192.168.1.100:8080/mcp')"],
      "env": {}
    }
  }
}
```

**多用户配置（支持不同用户使用不同账号）：**

从 v1.1.0 开始，服务器支持多用户配置，不同的客户端可以使用不同的 Confluence 账号：

```json
{
  "mcpServers": {
    "confluence": {
      "command": "npx",
      "args": ["@modelcontextprotocol/server-fetch"],
      "env": {
        "FETCH_BASE_URL": "http://192.168.1.100:8080/mcp",
        "FETCH_HEADERS": "X-Confluence-Base-URL=https://your-company.atlassian.net/wiki,X-Confluence-API-Token=your_api_token,X-Confluence-User-Email=your.email@company.com"
      }
    }
  }
}
```

**配置说明：**
- `192.168.1.100:8080` - 替换为运行 Confluence MCP 服务器的实际 IP 地址和端口
- 确保服务器已启动 HTTP 模式：`npm run server`
- 确保防火墙允许 8080 端口访问
- 可以通过 `http://192.168.1.100:8080/health` 测试连接
- **多用户支持**：每个客户端可以在请求头中提供自己的 Confluence 认证信息

#### 使用其他 MCP 客户端

```bash
# 直接运行 MCP 服务器
node dist/main.js

# 或者使用 npm 脚本
npm run start:mcp
```

### 2. 局域网部署配置

#### 服务器端配置

**1. 启动 HTTP 服务模式**

```bash
# 开发模式
npm run server

# 生产模式
npm run build
npm start
```

**2. 配置网络访问**

确保服务器监听所有网络接口（默认配置已支持）：

```bash
# 检查服务是否正常启动
curl http://localhost:8080/health

# 从其他机器测试连接（替换为实际 IP）
curl http://192.168.1.100:8080/health
```

**3. 防火墙配置**

```bash
# macOS 防火墙（如果启用）
sudo pfctl -f /etc/pf.conf

# Linux 防火墙
sudo ufw allow 8080

# Windows 防火墙
# 在 Windows 防火墙设置中允许端口 8080
```

#### 客户端配置

**获取服务器 IP 地址：**

```bash
# macOS/Linux
ifconfig | grep "inet " | grep -v 127.0.0.1

# 或者
ip addr show | grep "inet " | grep -v 127.0.0.1

# Windows
ipconfig
```

**测试连接：**

```bash
# 测试健康检查端点
curl http://[服务器IP]:8080/health

# 测试工具列表
curl http://[服务器IP]:8080/mcp/tools
```

### 3. 多用户配置管理

#### 概述

从 v1.1.0 开始，Confluence MCP 服务器支持多用户配置，允许不同的客户端使用不同的 Confluence 账号连接到同一个服务器实例。这对于团队环境特别有用，每个成员可以使用自己的 Confluence 账号进行操作。

#### 工作原理

服务器通过 HTTP 请求头接收用户特定的认证信息：
- `X-Confluence-Base-URL`: Confluence 实例的基础 URL
- `X-Confluence-API-Token`: 用户的 API 令牌
- `X-Confluence-User-Email`: 用户的邮箱地址

#### 配置方式

**方式一：通过 HTTP 客户端直接调用**

```javascript
// 用户 A 的配置
const userAConfig = {
  'X-Confluence-Base-URL': 'https://company.atlassian.net/wiki',
  'X-Confluence-API-Token': 'ATATT3xFfGF0T4JVjdmA1b2c3d4e5f6g7h8i9j0k',
  'X-Confluence-User-Email': 'alice@company.com'
};

// 用户 B 的配置
const userBConfig = {
  'X-Confluence-Base-URL': 'https://company.atlassian.net/wiki',
  'X-Confluence-API-Token': 'ATATT3xFfGF0T4JVjdmA2c3d4e5f6g7h8i9j0k1l',
  'X-Confluence-User-Email': 'bob@company.com'
};

// 调用 API
fetch('http://192.168.1.100:8080/mcp/call', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    ...userAConfig  // 使用用户 A 的配置
  },
  body: JSON.stringify({
    method: 'tools/call',
    params: {
      name: 'get_page',
      arguments: { pageId: '123456789' }
    }
  })
});
```

**方式二：通过环境变量（Claude Desktop 配置）**

每个用户可以配置自己的 Claude Desktop：

```json
{
  "mcpServers": {
    "confluence": {
      "command": "npx",
      "args": ["@modelcontextprotocol/server-fetch"],
      "env": {
        "FETCH_BASE_URL": "http://192.168.1.100:8080/mcp",
        "FETCH_HEADERS": "X-Confluence-Base-URL=https://company.atlassian.net/wiki,X-Confluence-API-Token=YOUR_PERSONAL_TOKEN,X-Confluence-User-Email=your.email@company.com"
      }
    }
  }
}
```

#### 用户管理端点

服务器提供了几个端点来管理和验证用户配置：

**1. 验证用户配置**
```bash
curl -X POST http://192.168.1.100:8080/auth/validate \
  -H "X-Confluence-Base-URL: https://company.atlassian.net/wiki" \
  -H "X-Confluence-API-Token: YOUR_TOKEN" \
  -H "X-Confluence-User-Email: your.email@company.com"
```

**2. 获取当前用户信息**
```bash
curl http://192.168.1.100:8080/auth/user \
  -H "X-Confluence-Base-URL: https://company.atlassian.net/wiki" \
  -H "X-Confluence-API-Token: YOUR_TOKEN" \
  -H "X-Confluence-User-Email: your.email@company.com"
```

**3. 健康检查（查看多用户支持状态）**
```bash
curl http://192.168.1.100:8080/health
```

#### 安全注意事项

1. **API 令牌安全**：确保 API 令牌的安全存储，不要在代码中硬编码
2. **HTTPS 传输**：在生产环境中使用 HTTPS 来保护认证信息
3. **访问控制**：配置适当的网络访问控制，限制服务器访问
4. **令牌轮换**：定期轮换 API 令牌

#### 故障排除

**认证失败：**
```bash
# 测试认证配置
curl -X POST http://192.168.1.100:8080/auth/validate \
  -H "X-Confluence-Base-URL: YOUR_BASE_URL" \
  -H "X-Confluence-API-Token: YOUR_TOKEN" \
  -H "X-Confluence-User-Email: YOUR_EMAIL"
```

**权限问题：**
- 确保 API 令牌有足够的权限
- 检查 Confluence 空间的访问权限
- 验证用户邮箱地址是否正确

### 4. SSE HTTP 连接（局域网访问）

#### JavaScript/TypeScript 客户端示例

```javascript
// 连接到 SSE 端点
const eventSource = new EventSource('http://localhost:8080/sse');

// 监听消息
eventSource.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('收到消息:', data);
  
  switch(data.type) {
    case 'connected':
      console.log('已连接到 MCP 服务器');
      break;
    case 'message':
      console.log('服务器消息:', data.content);
      break;
    case 'error':
      console.error('服务器错误:', data.content);
      break;
    case 'closed':
      console.log('连接已关闭');
      break;
  }
};

// 处理连接错误
eventSource.onerror = function(event) {
  console.error('SSE 连接错误:', event);
};

// 关闭连接
// eventSource.close();
```

#### HTTP API 调用示例

```javascript
// 获取可用工具列表
async function getTools() {
  const response = await fetch('http://localhost:8080/mcp/tools');
  const tools = await response.json();
  console.log('可用工具:', tools);
  return tools;
}

// 调用 MCP 工具
async function callTool(toolName, args) {
  const response = await fetch('http://localhost:8080/mcp/call', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      method: 'tools/call',
      params: {
        name: toolName,
        arguments: args
      }
    })
  });
  
  const result = await response.json();
  return result;
}

// 使用示例
async function example() {
  try {
    // 获取页面内容
    const pageResult = await callTool('get_page', {
      pageId: '123456789',
      includeComments: true
    });
    console.log('页面内容:', pageResult);
    
    // 搜索页面
    const searchResult = await callTool('search_pages', {
      query: 'API 文档',
      spaceKey: 'DEV',
      limit: 5
    });
    console.log('搜索结果:', searchResult);
    
  } catch (error) {
    console.error('调用失败:', error);
  }
}
```

#### Python 客户端示例

```python
import requests
import json
from sseclient import SSEClient

# SSE 连接
def connect_sse():
    url = 'http://localhost:8080/sse'
    messages = SSEClient(url)
    
    for msg in messages:
        if msg.data:
            data = json.loads(msg.data)
            print(f"收到消息: {data}")
            
            if data.get('type') == 'connected':
                print("已连接到 MCP 服务器")
            elif data.get('type') == 'error':
                print(f"服务器错误: {data.get('content')}")

# HTTP API 调用
def call_tool(tool_name, args):
    url = 'http://localhost:8080/mcp/call'
    payload = {
        'method': 'tools/call',
        'params': {
            'name': tool_name,
            'arguments': args
        }
    }
    
    response = requests.post(url, json=payload)
    return response.json()

# 使用示例
if __name__ == '__main__':
    # 获取页面内容
    result = call_tool('get_page', {
        'pageId': '123456789',
        'includeComments': True
    })
    print(f"页面内容: {result}")
```

#### cURL 命令示例

```bash
# 健康检查
curl http://localhost:8080/health

# 获取工具列表
curl http://localhost:8080/mcp/tools

# 调用工具 - 获取页面
curl -X POST http://localhost:8080/mcp/call \
  -H "Content-Type: application/json" \
  -d '{
    "method": "tools/call",
    "params": {
      "name": "get_page",
      "arguments": {
        "pageId": "123456789",
        "includeComments": true
      }
    }
  }'

# 调用工具 - 搜索页面
curl -X POST http://localhost:8080/mcp/call \
  -H "Content-Type: application/json" \
  -d '{
    "method": "tools/call",
    "params": {
      "name": "search_pages",
      "arguments": {
        "query": "API 文档",
        "spaceKey": "DEV",
        "limit": 5
      }
    }
  }'
```

### 3. 局域网访问配置

#### 获取服务器 IP 地址

```bash
# macOS/Linux
ifconfig | grep "inet " | grep -v 127.0.0.1

# 或者
ip addr show | grep "inet " | grep -v 127.0.0.1

# Windows
ipconfig | findstr "IPv4"
```

#### 局域网内其他设备访问

将 `localhost` 替换为服务器的实际 IP 地址：

```javascript
// 例如服务器 IP 为 192.168.1.100
const eventSource = new EventSource('http://192.168.1.100:8080/sse');

// HTTP API 调用
const response = await fetch('http://192.168.1.100:8080/mcp/tools');
```

#### 防火墙配置

确保防火墙允许 8080 端口的访问：

```bash
# macOS
sudo pfctl -f /etc/pf.conf

# Ubuntu/Debian
sudo ufw allow 8080

# CentOS/RHEL
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```
```

## 工具使用示例

### 1. 获取页面内容

```json
{
  "tool": "get_page",
  "arguments": {
    "pageId": "123456789",
    "includeComments": true
  }
}
```

**响应示例：**
```
# 页面信息

**标题:** API 开发指南
**ID:** 123456789
**空间:** 开发文档 (DEV)
**状态:** current
**版本:** 5
**最后修改:** 2024-01-15T10:30:00.000Z
**修改者:** 张三
**链接:** https://company.atlassian.net/wiki/spaces/DEV/pages/123456789

## 页面内容

<p>这里是页面的具体内容...</p>

## 评论 (2条)

### 评论 1
**创建者:** 李四
**创建时间:** 2024-01-14T15:20:00.000Z

很好的文档，建议添加更多示例。

### 评论 2
**创建者:** 王五
**创建时间:** 2024-01-15T09:15:00.000Z

已经按照这个文档完成了集成，非常有用！
```

### 2. 获取子页面列表

```json
{
  "tool": "get_child_pages",
  "arguments": {
    "pageId": "123456789",
    "limit": 10
  }
}
```

**响应示例：**
```
# 子页面列表

找到 3 个子页面 (父页面: API 开发指南)

| 标题                                                                              | ID        | 创建时间  | 创建者 | 最后修改时间 | 修改者 | 版本 | 状态    |
| --------------------------------------------------------------------------------- | --------- | --------- | ------ | ------------ | ------ | ---- | ------- |
| [REST API 规范](https://company.atlassian.net/wiki/spaces/DEV/pages/123456790)    | 123456790 | 2024/1/10 | 张三   | 2024/1/15    | 李四   | v3   | current |
| [GraphQL API 指南](https://company.atlassian.net/wiki/spaces/DEV/pages/123456791) | 123456791 | 2024/1/12 | 王五   | 2024/1/14    | 王五   | v2   | current |
| [API 测试指南](https://company.atlassian.net/wiki/spaces/DEV/pages/123456792)     | 123456792 | 2024/1/13 | 赵六   | 2024/1/13    | 赵六   | v1   | current |
```

### 3. 创建新页面

```json
{
  "tool": "create_page",
  "arguments": {
    "spaceKey": "DEV",
    "title": "新的 API 文档",
    "content": "<h1>新的 API 文档</h1><p>这是一个新创建的页面，用于记录最新的 API 变更。</p><ul><li>功能 A</li><li>功能 B</li></ul>",
    "parentId": "123456789"
  }
}
```

**响应示例：**
```
# 页面创建成功

✅ 新页面已成功创建！

## 页面信息

- **标题:** 新的 API 文档
- **页面ID:** 123456800
- **空间:** 开发文档 (DEV)
- **状态:** current
- **版本:** 1
- **创建时间:** 2024/1/16 14:30:25
- **创建者:** 张三 (zhang.san@company.com)
- **页面链接:** [点击查看](https://company.atlassian.net/wiki/spaces/DEV/pages/123456800)
- **父页面ID:** 123456789

## 页面内容预览

<h1>新的 API 文档</h1><p>这是一个新创建的页面，用于记录最新的 API 变更。</p><ul><li>功能 A</li><li>功能 B</li></ul>
```

### 4. 添加评论

```json
{
  "tool": "create_comment",
  "arguments": {
    "pageId": "123456789",
    "comment": "感谢分享这个文档！我有一个问题：关于认证部分，是否支持 OAuth 2.0？"
  }
}
```

**响应示例：**
```
# 评论创建成功

✅ 评论已成功添加到页面！

## 评论信息

- **评论ID:** 123456850
- **页面:** API 开发指南 (ID: 123456789)
- **状态:** current
- **版本:** 1
- **创建时间:** 2024/1/16 14:35:10
- **创建者:** 张三 (zhang.san@company.com)
- **评论链接:** [点击查看](https://company.atlassian.net/wiki/spaces/DEV/pages/123456789#comment-123456850)

## 评论内容

感谢分享这个文档！我有一个问题：关于认证部分，是否支持 OAuth 2.0？
```

### 5. 搜索页面

```json
{
  "tool": "search_pages",
  "arguments": {
    "query": "API 文档",
    "spaceKey": "DEV",
    "limit": 5
  }
}
```

**响应示例：**
```
# 搜索结果

搜索关键字: "API 文档" (限制在空间: DEV)
找到 8 个结果，显示前 5 个

## 故障排除

### 局域网连接问题

#### 1. 无法连接到服务器

**检查服务器状态：**
```bash
# 在服务器机器上检查服务是否运行
curl http://localhost:8080/health

# 检查端口是否被占用
lsof -i :8080  # macOS/Linux
netstat -ano | findstr :8080  # Windows
```

**检查网络连通性：**
```bash
# 从客户端机器 ping 服务器
ping 192.168.1.100

# 检查端口是否可达
telnet 192.168.1.100 8080
# 或者使用 nc
nc -zv 192.168.1.100 8080
```

#### 2. 防火墙问题

**macOS 防火墙：**
```bash
# 检查防火墙状态
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate

# 临时关闭防火墙测试
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --setglobalstate off
```

**Linux 防火墙：**
```bash
# Ubuntu/Debian
sudo ufw status
sudo ufw allow 8080

# CentOS/RHEL
sudo firewall-cmd --list-ports
sudo firewall-cmd --add-port=8080/tcp --permanent
sudo firewall-cmd --reload
```

#### 3. Claude Desktop 配置问题

**常见错误和解决方案：**

1. **"Command not found" 错误：**
   - 确保安装了 `@modelcontextprotocol/server-fetch`：
     ```bash
     npm install -g @modelcontextprotocol/server-fetch
     ```

2. **"Connection refused" 错误：**
   - 检查 URL 是否正确（包括协议、IP、端口）
   - 确保服务器正在运行 HTTP 模式
   - 验证网络连通性

3. **"Timeout" 错误：**
   - 检查网络延迟
   - 增加超时设置
   - 确保服务器性能正常

#### 4. 网络配置检查

**获取正确的 IP 地址：**
```bash
# 显示所有网络接口
ifconfig -a  # macOS/Linux
ipconfig /all  # Windows

# 只显示活动的 IPv4 地址
ifconfig | grep "inet " | grep -v "127.0.0.1" | awk '{print $2}'  # macOS/Linux
```

**测试不同端口：**
```bash
# 如果 8080 被占用，可以修改端口
export PORT=8081
npm run server
```

#### 5. Docker 部署问题

**端口映射检查：**
```bash
# 检查容器端口映射
docker ps

# 检查容器日志
docker logs confluence-mcp

# 进入容器调试
docker exec -it confluence-mcp sh
```

**网络模式配置：**
```yaml
# docker-compose.yml
services:
  confluence-mcp:
    ports:
      - "8080:8080"
    networks:
      - bridge
```

### 性能优化建议

1. **使用有线网络连接**以获得更稳定的性能
2. **配置静态 IP 地址**避免 DHCP 变更导致的连接问题
3. **使用反向代理**（如 Nginx）进行负载均衡和 SSL 终止
4. **启用 HTTP/2**提升连接性能
5. **配置适当的超时设置**适应网络环境

### 日志和调试

**启用详细日志：**
```bash
# 设置日志级别
export LOG_LEVEL=debug
npm run server

# 或者在 .env 文件中设置
echo "LOG_LEVEL=debug" >> .env
```

**查看连接日志：**
```bash
# 实时查看日志
tail -f logs/app.log

# 过滤特定类型的日志
grep "ERROR" logs/app.log
grep "connection" logs/app.log
```

## 搜索结果列表

### 1. [API 开发指南](https://company.atlassian.net/wiki/spaces/DEV/pages/123456789)

**页面信息:**
- **ID:** 123456789
- **空间:** 开发文档 (DEV)
- **状态:** current
- **创建时间:** 2024/1/10 by 张三
- **最后修改:** 2024/1/15 by 李四
- **版本:** v5

**内容摘要:**
这是一份完整的 API 开发指南，包含了所有必要的信息来帮助开发者快速上手我们的 API。本文档涵盖了认证、请求格式、响应格式、错误处理等关键主题...

---

### 2. [REST API 规范](https://company.atlassian.net/wiki/spaces/DEV/pages/123456790)

**页面信息:**
- **ID:** 123456790
- **空间:** 开发文档 (DEV)
- **状态:** current
- **创建时间:** 2024/1/10 by 张三
- **最后修改:** 2024/1/15 by 李四
- **版本:** v3

**内容摘要:**
REST API 的详细规范文档，定义了所有端点的请求和响应格式。包括用户管理、数据查询、文件上传等功能的 API 接口说明...

---

💡 **提示:** 还有 3 个结果未显示。如需查看更多结果，请增加 limit 参数或使用更具体的搜索关键字。
```

## 常见问题

### Q: 如何获取页面ID？
A: 在 Confluence 页面中，页面ID通常出现在URL中，格式如：`/pages/123456789/page-title`

### Q: 支持哪些内容格式？
A: 支持 Confluence 存储格式（Storage Format），这是一种基于 XHTML 的格式。

### Q: 如何处理权限问题？
A: 确保你的 API 令牌对应的用户有足够的权限访问和修改相关页面和空间。

### Q: 搜索支持哪些语法？
A: 使用 Confluence Query Language (CQL) 语法，支持文本搜索、字段过滤等。

## 故障排除

### 认证失败
- 检查 API 令牌是否正确
- 确认用户邮箱是否正确
- 验证 Confluence 实例 URL 是否正确

### 权限错误
- 确保用户有访问指定空间的权限
- 检查是否有创建/编辑页面的权限
- 验证是否有添加评论的权限

### 页面不存在
- 确认页面ID是否正确
- 检查页面是否已被删除
- 验证是否有查看该页面的权限