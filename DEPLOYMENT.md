# Confluence MCP 服务器 - SSE 部署指南

本指南介绍如何将 Confluence MCP 服务器部署为 SSE (Server-Sent Events) 服务，供局域网使用。

## 🚀 快速部署

### 1. 环境准备

确保已安装 Docker 和 Docker Compose：

```bash
# 检查 Docker 版本
docker --version
docker-compose --version
```

### 2. 配置环境变量

复制环境变量模板并配置：

```bash
cp .env.example .env
```

编辑 `.env` 文件，配置 Confluence API 信息：

```env
# Confluence API 配置
CONFLUENCE_BASE_URL=https://your-domain.atlassian.net
CONFLUENCE_API_TOKEN=your_api_token_here
CONFLUENCE_USER_EMAIL=your-email@example.com

# MCP 服务器配置
MCP_SERVER_NAME=confluence-mcp
MCP_SERVER_VERSION=1.0.0
LOG_LEVEL=info
NODE_ENV=production
```

### 3. 构建和启动服务

```bash
# 构建并启动服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f confluence-mcp
```

### 4. 验证部署

服务启动后，可以通过以下端点验证：

```bash
# 健康检查
curl http://localhost:8080/health

# 获取可用工具列表
curl http://localhost:8080/mcp/tools
```

## 📡 SSE 接口使用

### SSE 连接

```javascript
// 连接到 SSE 端点
const eventSource = new EventSource('http://localhost:8080/sse');

eventSource.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('收到消息:', data);
};

eventSource.onerror = function(event) {
  console.error('SSE 连接错误:', event);
};
```

### HTTP API 调用

```bash
# 调用 MCP 工具
curl -X POST http://localhost:8080/mcp/call \
  -H "Content-Type: application/json" \
  -d '{
    "method": "tools/call",
    "params": {
      "name": "get_page",
      "arguments": {
        "pageId": "123456",
        "includeComments": true
      }
    }
  }'
```

## 🌐 局域网访问配置

### 1. 修改端口映射（可选）

如果需要使用其他端口，修改 `docker-compose.yml`：

```yaml
ports:
  - "9090:8080"  # 将本地 9090 端口映射到容器 8080 端口
```

### 2. 防火墙配置

确保防火墙允许访问指定端口：

```bash
# Ubuntu/Debian
sudo ufw allow 8080

# CentOS/RHEL
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload

# macOS
# 在系统偏好设置 > 安全性与隐私 > 防火墙中配置
```

### 3. 获取服务器 IP 地址

```bash
# Linux/macOS
ip addr show | grep inet
# 或
ifconfig | grep inet

# Windows
ipconfig
```

局域网内其他设备可通过 `http://服务器IP:8080` 访问服务。

## 🔧 高级配置

### 自定义网络配置

修改 `docker-compose.yml` 中的网络配置：

```yaml
networks:
  mcp-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
          gateway: 172.20.0.1
```

### 资源限制

为容器设置资源限制：

```yaml
confluence-mcp:
  # ... 其他配置
  deploy:
    resources:
      limits:
        cpus: '0.5'
        memory: 512M
      reservations:
        cpus: '0.25'
        memory: 256M
```

### 持久化日志

日志会自动保存到 `./logs` 目录，可以通过以下方式查看：

```bash
# 查看实时日志
tail -f logs/app.log

# 查看容器日志
docker-compose logs -f confluence-mcp
```

## 🛠️ 故障排除

### 常见问题

1. **容器启动失败**
   ```bash
   # 检查容器状态
   docker-compose ps
   
   # 查看详细日志
   docker-compose logs confluence-mcp
   ```

2. **端口被占用**
   ```bash
   # 检查端口占用
   lsof -i :8080
   
   # 修改 docker-compose.yml 中的端口映射
   ports:
     - "8081:8080"
   ```

3. **Confluence API 连接失败**
   - 检查 `.env` 文件中的配置
   - 验证 API Token 是否有效
   - 确认网络连接正常

4. **健康检查失败**
   ```bash
   # 手动测试健康检查
   docker exec confluence-mcp-server wget --spider http://localhost:8080/health
   ```

### 重新部署

```bash
# 停止服务
docker-compose down

# 重新构建并启动
docker-compose up -d --build

# 清理旧镜像（可选）
docker image prune -f
```

## 📊 监控和维护

### 服务监控

```bash
# 查看容器资源使用情况
docker stats confluence-mcp-server

# 查看容器详细信息
docker inspect confluence-mcp-server
```

### 日志轮转

配置 Docker 日志轮转：

```yaml
confluence-mcp:
  # ... 其他配置
  logging:
    driver: "json-file"
    options:
      max-size: "10m"
      max-file: "3"
```

### 备份和恢复

```bash
# 备份配置文件
tar -czf confluence-mcp-backup.tar.gz .env docker-compose.yml

# 恢复配置
tar -xzf confluence-mcp-backup.tar.gz
```

## 🔒 安全建议

1. **API Token 安全**
   - 使用具有最小权限的 API Token
   - 定期轮换 API Token
   - 不要在代码中硬编码敏感信息

2. **网络安全**
   - 在生产环境中使用 HTTPS
   - 配置适当的防火墙规则
   - 考虑使用 VPN 或内网访问

3. **容器安全**
   - 定期更新基础镜像
   - 使用非 root 用户运行容器
   - 限制容器资源使用

## 📞 支持

如果遇到问题，请检查：

1. [README.md](./README.md) - 基本使用说明
2. [example-usage.md](./example-usage.md) - 使用示例
3. 容器日志和应用日志
4. Confluence API 文档

---

**注意**: 本部署方案适用于开发和测试环境。生产环境部署请考虑额外的安全措施、负载均衡、监控和备份策略。