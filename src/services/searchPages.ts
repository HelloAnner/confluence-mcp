import axios from 'axios';

// Confluence API 配置
const CONFLUENCE_BASE_URL = process.env.CONFLUENCE_BASE_URL!;
const CONFLUENCE_API_TOKEN = process.env.CONFLUENCE_API_TOKEN!;
const CONFLUENCE_USER_EMAIL = process.env.CONFLUENCE_USER_EMAIL!;

// 创建 axios 实例
const confluenceApi = axios.create({
  baseURL: `${CONFLUENCE_BASE_URL}/rest/api`,
  auth: {
    username: CONFLUENCE_USER_EMAIL,
    password: CONFLUENCE_API_TOKEN
  },
  headers: {
    'Accept': 'application/json',
    'Content-Type': 'application/json'
  }
});

// 搜索参数接口
interface SearchParams {
  query: string;
  spaceKey?: string;
  limit?: number;
}

// 搜索结果接口
interface SearchResult {
  id: string;
  title: string;
  type: string;
  status: string;
  space: {
    key: string;
    name: string;
  };
  version: {
    number: number;
    when: string;
    by: {
      displayName: string;
      email: string;
    };
  };
  history: {
    createdDate: string;
    createdBy: {
      displayName: string;
      email: string;
    };
  };
  body?: {
    storage: {
      value: string;
      representation: string;
    };
  };
  excerpt?: string;
  _links: {
    webui: string;
  };
}

/**
 * 根据关键字搜索Confluence页面
 * @param params 搜索参数
 * @returns 搜索结果列表
 */
export async function searchPages(params: SearchParams) {
  const { query, spaceKey, limit = 20 } = params;

  try {
    // 构建搜索查询
    let cqlQuery = `type=page AND text~"${query}"`;
    
    // 如果指定了空间，添加空间限制
    if (spaceKey) {
      // 验证空间是否存在
      try {
        await confluenceApi.get(`/space/${spaceKey}`);
        cqlQuery += ` AND space="${spaceKey}"`;
      } catch (spaceError) {
        if (axios.isAxiosError(spaceError) && spaceError.response?.status === 404) {
          throw new Error(`空间不存在: ${spaceKey}`);
        }
        throw spaceError;
      }
    }

    // 执行搜索
    const response = await confluenceApi.get('/content/search', {
      params: {
        cql: cqlQuery,
        expand: 'space,version,history,body.storage',
        limit: Math.min(limit, 100), // 限制最大值为100
        start: 0
      }
    });

    const searchResults: SearchResult[] = response.data.results;
    const totalSize = response.data.size;

    // 格式化搜索结果
    const formattedResults = searchResults.map(result => {
      // 提取内容摘要
      let excerpt = '';
      if (result.body && result.body.storage) {
        const content = result.body.storage.value;
        // 移除HTML标签并截取前200个字符
        const plainText = content.replace(/<[^>]*>/g, '').replace(/\s+/g, ' ').trim();
        excerpt = plainText.length > 200 ? plainText.substring(0, 200) + '...' : plainText;
      }

      return {
        id: result.id,
        title: result.title,
        type: result.type,
        status: result.status,
        space: {
          key: result.space.key,
          name: result.space.name
        },
        createdAt: result.history.createdDate,
        createdBy: {
          name: result.history.createdBy.displayName,
          email: result.history.createdBy.email
        },
        lastModified: result.version.when,
        lastModifiedBy: {
          name: result.version.by.displayName,
          email: result.version.by.email
        },
        version: result.version.number,
        excerpt: excerpt,
        webUrl: `${CONFLUENCE_BASE_URL}${result._links.webui}`
      };
    });

    // 构建返回结果
    const searchSummary = `搜索关键字: "${query}"${spaceKey ? ` (限制在空间: ${spaceKey})` : ''}\n找到 ${totalSize} 个结果${totalSize > limit ? `，显示前 ${formattedResults.length} 个` : ''}`;
    
    let resultText = `# 搜索结果\n\n${searchSummary}\n\n`;
    
    if (formattedResults.length === 0) {
      resultText += '没有找到匹配的页面。\n\n**搜索建议:**\n- 尝试使用不同的关键字\n- 检查拼写是否正确\n- 使用更通用的搜索词';
    } else {
      resultText += '## 搜索结果列表\n\n';
      
      formattedResults.forEach((result, index) => {
        const createdDate = new Date(result.createdAt).toLocaleDateString('zh-CN');
        const modifiedDate = new Date(result.lastModified).toLocaleDateString('zh-CN');
        
        resultText += `### ${index + 1}. [${result.title}](${result.webUrl})\n\n`;
        resultText += `**页面信息:**\n`;
        resultText += `- **ID:** ${result.id}\n`;
        resultText += `- **空间:** ${result.space.name} (${result.space.key})\n`;
        resultText += `- **状态:** ${result.status}\n`;
        resultText += `- **创建时间:** ${createdDate} by ${result.createdBy.name}\n`;
        resultText += `- **最后修改:** ${modifiedDate} by ${result.lastModifiedBy.name}\n`;
        resultText += `- **版本:** v${result.version}\n\n`;
        
        if (result.excerpt) {
          resultText += `**内容摘要:**\n${result.excerpt}\n\n`;
        }
        
        resultText += '---\n\n';
      });
      
      if (totalSize > limit) {
        resultText += `\n💡 **提示:** 还有 ${totalSize - formattedResults.length} 个结果未显示。如需查看更多结果，请增加 limit 参数或使用更具体的搜索关键字。`;
      }
    }

    return {
      content: [
        {
          type: 'text',
          text: resultText
        }
      ]
    };

  } catch (error) {
    console.error('搜索页面失败:', error);
    
    if (axios.isAxiosError(error)) {
      const status = error.response?.status;
      const message = error.response?.data?.message || error.message;
      
      if (status === 400) {
        throw new Error(`搜索查询无效: ${message}`);
      } else if (status === 403) {
        throw new Error(`没有权限执行搜索${spaceKey ? ` (空间: ${spaceKey})` : ''}`);
      } else if (status === 401) {
        throw new Error('认证失败，请检查API令牌和用户邮箱');
      } else {
        throw new Error(`搜索失败: ${message}`);
      }
    }
    
    throw new Error(`搜索时发生未知错误: ${error}`);
  }
}