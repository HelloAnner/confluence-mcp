import { spawn } from 'child_process';
import cors from 'cors';
import express from 'express';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const app = express();
const PORT = process.env.PORT || 8080;

// 启用CORS以支持跨域请求
app.use(cors({
  origin: '*',
  methods: ['GET', 'POST', 'OPTIONS'],
  allowedHeaders: ['Content-Type', 'Authorization', 'Cache-Control']
}));

app.use(express.json());

// 健康检查端点
app.get('/health', (req, res) => {
  res.json({ 
    status: 'ok', 
    service: 'confluence-mcp-server',
    version: process.env.MCP_SERVER_VERSION || '1.0.0',
    timestamp: new Date().toISOString()
  });
});

// SSE端点
app.get('/sse', (req, res) => {
  // 设置SSE头
  res.writeHead(200, {
    'Content-Type': 'text/event-stream',
    'Cache-Control': 'no-cache',
    'Connection': 'keep-alive',
    'Access-Control-Allow-Origin': '*',
    'Access-Control-Allow-Headers': 'Cache-Control'
  });

  // 发送初始连接消息
  res.write(`data: ${JSON.stringify({ type: 'connected', message: 'MCP Server connected' })}\n\n`);

  // 启动MCP服务器进程
  const mcpProcess = spawn('node', [path.join(__dirname, 'main.js')], {
    stdio: ['pipe', 'pipe', 'pipe']
  });

  // 处理MCP服务器输出
  mcpProcess.stdout.on('data', (data) => {
    const output = data.toString();
    try {
      // 尝试解析JSON响应
      const jsonData = JSON.parse(output);
      res.write(`data: ${JSON.stringify(jsonData)}\n\n`);
    } catch (e) {
      // 如果不是JSON，作为普通消息发送
      res.write(`data: ${JSON.stringify({ type: 'message', content: output })}\n\n`);
    }
  });

  mcpProcess.stderr.on('data', (data) => {
    const error = data.toString();
    res.write(`data: ${JSON.stringify({ type: 'error', content: error })}\n\n`);
  });

  mcpProcess.on('close', (code) => {
    res.write(`data: ${JSON.stringify({ type: 'closed', code })}\n\n`);
    res.end();
  });

  // 处理客户端断开连接
  req.on('close', () => {
    mcpProcess.kill();
  });
});

// MCP工具调用端点
app.post('/mcp/call', async (req, res) => {
  try {
    const { method, params } = req.body;
    
    // 启动MCP服务器进程
    const mcpProcess = spawn('node', [path.join(__dirname, 'main.js')], {
      stdio: ['pipe', 'pipe', 'pipe']
    });

    let responseData = '';
    let errorData = '';

    // 发送请求到MCP服务器
    mcpProcess.stdin.write(JSON.stringify({
      jsonrpc: '2.0',
      id: Date.now(),
      method,
      params
    }) + '\n');
    mcpProcess.stdin.end();

    // 收集响应
    mcpProcess.stdout.on('data', (data) => {
      responseData += data.toString();
    });

    mcpProcess.stderr.on('data', (data) => {
      errorData += data.toString();
    });

    mcpProcess.on('close', (code) => {
      if (code === 0 && responseData) {
        try {
          const jsonResponse = JSON.parse(responseData);
          res.json(jsonResponse);
        } catch (e) {
          res.status(500).json({ error: 'Invalid JSON response from MCP server' });
        }
      } else {
        res.status(500).json({ 
          error: 'MCP server error', 
          code, 
          stderr: errorData 
        });
      }
    });

  } catch (error) {
    res.status(500).json({ 
      error: 'Internal server error', 
      message: error instanceof Error ? error.message : String(error)
    });
  }
});

// 获取可用工具列表
app.get('/mcp/tools', async (req, res) => {
  try {
    const mcpProcess = spawn('node', [path.join(__dirname, 'main.js')], {
      stdio: ['pipe', 'pipe', 'pipe']
    });

    let responseData = '';

    // 请求工具列表
    mcpProcess.stdin.write(JSON.stringify({
      jsonrpc: '2.0',
      id: 1,
      method: 'tools/list'
    }) + '\n');
    mcpProcess.stdin.end();

    mcpProcess.stdout.on('data', (data) => {
      responseData += data.toString();
    });

    mcpProcess.on('close', (code) => {
      if (code === 0 && responseData) {
        try {
          const jsonResponse = JSON.parse(responseData);
          res.json(jsonResponse);
        } catch (e) {
          res.status(500).json({ error: 'Invalid JSON response' });
        }
      } else {
        res.status(500).json({ error: 'Failed to get tools list' });
      }
    });

  } catch (error) {
    res.status(500).json({ 
      error: 'Internal server error', 
      message: error instanceof Error ? error.message : String(error)
    });
  }
});

// 启动服务器
app.listen(PORT as number, '0.0.0.0', () => {
  console.log(`🚀 Confluence MCP Server HTTP wrapper running on port ${PORT}`);
  console.log(`📡 SSE endpoint: http://localhost:${PORT}/sse`);
  console.log(`🔧 Tools endpoint: http://localhost:${PORT}/mcp/tools`);
  console.log(`💬 Call endpoint: http://localhost:${PORT}/mcp/call`);
  console.log(`❤️  Health check: http://localhost:${PORT}/health`);
});

export default app;