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

// 子页面信息接口
interface ChildPageInfo {
  id: string;
  title: string;
  type: string;
  status: string;
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
  _links: {
    webui: string;
  };
}

/**
 * 获取指定页面的子页面列表信息
 * @param pageId 父页面ID
 * @param limit 返回结果的最大数量
 * @returns 子页面列表信息
 */
export async function getChildPages(pageId: string, limit: number = 50) {
  try {
    // 获取子页面列表
    const response = await confluenceApi.get(`/content/${pageId}/child/page`, {
      params: {
        expand: 'version,history,space',
        limit: Math.min(limit, 200) // 限制最大值为200
      }
    });

    const childPages: ChildPageInfo[] = response.data.results;

    // 格式化子页面信息
    const formattedPages = childPages.map(page => ({
      id: page.id,
      title: page.title,
      type: page.type,
      status: page.status,
      createdAt: page.history.createdDate,
      createdBy: {
        name: page.history.createdBy.displayName,
        email: page.history.createdBy.email
      },
      lastModified: page.version.when,
      lastModifiedBy: {
        name: page.version.by.displayName,
        email: page.version.by.email
      },
      version: page.version.number,
      webUrl: `${CONFLUENCE_BASE_URL}${page._links.webui}`
    }));

    // 获取父页面信息
    let parentPageTitle = '';
    try {
      const parentResponse = await confluenceApi.get(`/content/${pageId}`, {
        params: {
          expand: 'space'
        }
      });
      parentPageTitle = parentResponse.data.title;
    } catch (error) {
      console.warn(`获取父页面 ${pageId} 信息时出错:`, error);
    }

    // 构建返回结果
    const summary = `找到 ${formattedPages.length} 个子页面${parentPageTitle ? ` (父页面: ${parentPageTitle})` : ''}`;
    
    let resultText = `# 子页面列表\n\n${summary}\n\n`;
    
    if (formattedPages.length === 0) {
      resultText += '暂无子页面。';
    } else {
      resultText += '| 标题 | ID | 创建时间 | 创建者 | 最后修改时间 | 修改者 | 版本 | 状态 |\n';
      resultText += '|------|----|---------|----|------------|----|----|------|\n';
      
      formattedPages.forEach(page => {
        const createdDate = new Date(page.createdAt).toLocaleDateString('zh-CN');
        const modifiedDate = new Date(page.lastModified).toLocaleDateString('zh-CN');
        
        resultText += `| [${page.title}](${page.webUrl}) | ${page.id} | ${createdDate} | ${page.createdBy.name} | ${modifiedDate} | ${page.lastModifiedBy.name} | v${page.version} | ${page.status} |\n`;
      });
      
      resultText += `\n\n## 详细信息\n\n`;
      
      formattedPages.forEach((page, index) => {
        resultText += `### ${index + 1}. ${page.title}\n`;
        resultText += `- **页面ID:** ${page.id}\n`;
        resultText += `- **创建时间:** ${new Date(page.createdAt).toLocaleString('zh-CN')}\n`;
        resultText += `- **创建者:** ${page.createdBy.name} (${page.createdBy.email})\n`;
        resultText += `- **最后修改:** ${new Date(page.lastModified).toLocaleString('zh-CN')}\n`;
        resultText += `- **修改者:** ${page.lastModifiedBy.name} (${page.lastModifiedBy.email})\n`;
        resultText += `- **版本:** ${page.version}\n`;
        resultText += `- **状态:** ${page.status}\n`;
        resultText += `- **链接:** ${page.webUrl}\n\n`;
      });
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
    console.error('获取子页面列表失败:', error);
    
    if (axios.isAxiosError(error)) {
      const status = error.response?.status;
      const message = error.response?.data?.message || error.message;
      
      if (status === 404) {
        throw new Error(`父页面不存在 (ID: ${pageId})`);
      } else if (status === 403) {
        throw new Error(`没有权限访问页面 (ID: ${pageId})`);
      } else if (status === 401) {
        throw new Error('认证失败，请检查API令牌和用户邮箱');
      } else {
        throw new Error(`获取子页面列表失败: ${message}`);
      }
    }
    
    throw new Error(`获取子页面列表时发生未知错误: ${error}`);
  }
}