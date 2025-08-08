import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  CallToolRequestSchema,
  ErrorCode,
  ListToolsRequestSchema,
  McpError,
} from '@modelcontextprotocol/sdk/types.js';
import dotenv from 'dotenv';

// 导入功能模块
import { createComment } from './services/createComment';
import { createPage } from './services/createPage';
import { getChildPages } from './services/getChildPages';
import { getPageById } from './services/getPage';
import { searchPages } from './services/searchPages';

// 加载环境变量
dotenv.config();

// 验证必需的环境变量
if (!process.env.CONFLUENCE_BASE_URL || !process.env.CONFLUENCE_API_TOKEN || !process.env.CONFLUENCE_USER_EMAIL) {
  console.error('❌ 缺少必需的环境变量:');
  console.error('   CONFLUENCE_BASE_URL - Confluence 实例的基础 URL');
  console.error('   CONFLUENCE_API_TOKEN - Confluence API 令牌');
  console.error('   CONFLUENCE_USER_EMAIL - Confluence 用户邮箱');
  process.exit(1);
}

// 创建 MCP 服务器实例
const server = new Server(
  {
    name: 'confluence-mcp',
    version: '1.0.0',
  },
  {
    capabilities: {
      tools: {},
    },
  }
);

// 定义可用的工具
const tools = [
  {
    name: 'get_page',
    description: '根据页面ID获取Confluence页面内容和评论',
    inputSchema: {
      type: 'object',
      properties: {
        pageId: {
          type: 'string',
          description: 'Confluence页面的ID'
        },
        includeComments: {
          type: 'boolean',
          description: '是否包含评论信息',
          default: true
        }
      },
      required: ['pageId']
    }
  },
  {
    name: 'get_child_pages',
    description: '获取指定页面的子页面列表信息',
    inputSchema: {
      type: 'object',
      properties: {
        pageId: {
          type: 'string',
          description: '父页面的ID'
        },
        limit: {
          type: 'number',
          description: '返回结果的最大数量',
          default: 50
        }
      },
      required: ['pageId']
    }
  },
  {
    name: 'create_page',
    description: '创建新的Confluence页面',
    inputSchema: {
      type: 'object',
      properties: {
        spaceKey: {
          type: 'string',
          description: '空间键值',
          default: ''
        },
        title: {
          type: 'string',
          description: '页面标题'
        },
        content: {
          type: 'string',
          description: '页面内容（支持Confluence存储格式）'
        },
        parentId: {
          type: 'string',
          description: '父页面ID（可选）'
        }
      },
      required: ['spaceKey', 'title', 'content']
    }
  },
  {
    name: 'create_comment',
    description: '为指定页面创建评论',
    inputSchema: {
      type: 'object',
      properties: {
        pageId: {
          type: 'string',
          description: '页面ID'
        },
        comment: {
          type: 'string',
          description: '评论内容'
        }
      },
      required: ['pageId', 'comment']
    }
  },
  {
    name: 'search_pages',
    description: '根据关键字搜索Confluence页面',
    inputSchema: {
      type: 'object',
      properties: {
        query: {
          type: 'string',
          description: '搜索关键字'
        },
        spaceKey: {
          type: 'string',
          description: '限制搜索的空间（可选）'
        },
        limit: {
          type: 'number',
          description: '返回结果的最大数量',
          default: 20
        }
      },
      required: ['query']
    }
  }
];

// 注册工具列表处理器
server.setRequestHandler(ListToolsRequestSchema, async () => {
  return {
    tools
  };
});

// 注册工具调用处理器
server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;

  // 确保 args 存在
  if (!args) {
    throw new McpError(
      ErrorCode.InvalidParams,
      '缺少必需的参数'
    );
  }

  try {
    switch (name) {
      case 'get_page':
        return await getPageById(
          args.pageId as string,
          (args.includeComments as boolean) ?? true
        );

      case 'get_child_pages':
        return await getChildPages(
          args.pageId as string,
          (args.limit as number) ?? 50
        );

      case 'create_page':
        return await createPage({
          spaceKey: args.spaceKey as string,
          title: args.title as string,
          content: args.content as string,
          parentId: args.parentId as string | undefined
        });

      case 'create_comment':
        return await createComment(
          args.pageId as string,
          args.comment as string
        );

      case 'search_pages':
        return await searchPages({
          query: args.query as string,
          spaceKey: args.spaceKey as string | undefined,
          limit: (args.limit as number) ?? 20
        });

      default:
        throw new McpError(
          ErrorCode.MethodNotFound,
          `未知的工具: ${name}`
        );
    }
  } catch (error) {
    console.error(`执行工具 ${name} 时出错:`, error);

    if (error instanceof McpError) {
      throw error;
    }

    throw new McpError(
      ErrorCode.InternalError,
      `执行工具时发生错误: ${error instanceof Error ? error.message : String(error)}`
    );
  }
});

// 启动服务器
async function main() {
  const transport = new StdioServerTransport();
  await server.connect(transport);

  console.error('🚀 Confluence MCP 服务器已启动');
  console.error('📋 可用工具:');
  tools.forEach(tool => {
    console.error(`   - ${tool.name}: ${tool.description}`);
  });
}

// 错误处理
process.on('SIGINT', async () => {
  console.error('\n👋 正在关闭服务器...');
  await server.close();
  process.exit(0);
});

process.on('unhandledRejection', (reason, promise) => {
  console.error('未处理的 Promise 拒绝:', reason);
});

process.on('uncaughtException', (error) => {
  console.error('未捕获的异常:', error);
  process.exit(1);
});

// 启动主函数
main().catch((error) => {
  console.error('启动服务器失败:', error);
  process.exit(1);
});
