# 使用官方 Node.js 18 Alpine 镜像作为基础镜像
FROM node:18-alpine

# 设置工作目录
WORKDIR /app

# 复制 package.json 和 package-lock.json
COPY package*.json ./

# 安装所有依赖（包括开发依赖，用于构建）
RUN npm ci

# 复制源代码
COPY . .

# 构建应用
RUN npm run build

# 清理开发依赖，只保留生产依赖
RUN npm ci --only=production && npm cache clean --force

# 创建非 root 用户
RUN addgroup -g 1001 -S nodejs
RUN adduser -S mcpuser -u 1001

# 更改文件所有权
RUN chown -R mcpuser:nodejs /app
USER mcpuser

# 暴露端口（用于SSE HTTP接口）
EXPOSE 8080

# 设置环境变量
ENV NODE_ENV=production
ENV MCP_SERVER_NAME=confluence-mcp
ENV MCP_SERVER_VERSION=1.0.0

# 启动MCP服务器
CMD ["npm", "start"]