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

// 创建页面参数接口
interface CreatePageParams {
  spaceKey: string;
  title: string;
  content: string;
  parentId?: string;
}

// 页面创建响应接口
interface CreatePageResponse {
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
  _links: {
    webui: string;
  };
}

/**
 * 创建新的Confluence页面
 * @param params 创建页面的参数
 * @returns 创建的页面信息
 */
export async function createPage(params: CreatePageParams) {
  const { spaceKey, title, content, parentId } = params;

  try {
    // 验证空间是否存在
    try {
      await confluenceApi.get(`/space/${spaceKey}`);
    } catch (spaceError) {
      if (axios.isAxiosError(spaceError) && spaceError.response?.status === 404) {
        throw new Error(`空间不存在: ${spaceKey}`);
      }
      throw spaceError;
    }

    // 如果指定了父页面，验证父页面是否存在
    if (parentId) {
      try {
        await confluenceApi.get(`/content/${parentId}`);
      } catch (parentError) {
        if (axios.isAxiosError(parentError) && parentError.response?.status === 404) {
          throw new Error(`父页面不存在: ${parentId}`);
        }
        throw parentError;
      }
    }

    // 构建页面数据
    const pageData: any = {
      type: 'page',
      title: title,
      space: {
        key: spaceKey
      },
      body: {
        storage: {
          value: content,
          representation: 'storage'
        }
      }
    };

    // 如果有父页面，添加父页面信息
    if (parentId) {
      pageData.ancestors = [
        {
          id: parentId
        }
      ];
    }

    // 创建页面
    const response = await confluenceApi.post('/content', pageData);
    const createdPage: CreatePageResponse = response.data;

    // 格式化返回结果
    const pageInfo = {
      id: createdPage.id,
      title: createdPage.title,
      type: createdPage.type,
      status: createdPage.status,
      space: {
        key: createdPage.space.key,
        name: createdPage.space.name
      },
      version: {
        number: createdPage.version.number,
        createdAt: createdPage.version.when,
        createdBy: {
          name: createdPage.version.by.displayName,
          email: createdPage.version.by.email
        }
      },
      webUrl: `${CONFLUENCE_BASE_URL}${createdPage._links.webui}`
    };

    const resultText = `# 页面创建成功\n\n` +
      `✅ 新页面已成功创建！\n\n` +
      `## 页面信息\n\n` +
      `- **标题:** ${pageInfo.title}\n` +
      `- **页面ID:** ${pageInfo.id}\n` +
      `- **空间:** ${pageInfo.space.name} (${pageInfo.space.key})\n` +
      `- **状态:** ${pageInfo.status}\n` +
      `- **版本:** ${pageInfo.version.number}\n` +
      `- **创建时间:** ${new Date(pageInfo.version.createdAt).toLocaleString('zh-CN')}\n` +
      `- **创建者:** ${pageInfo.version.createdBy.name} (${pageInfo.version.createdBy.email})\n` +
      `- **页面链接:** [点击查看](${pageInfo.webUrl})\n\n` +
      `${parentId ? `- **父页面ID:** ${parentId}\n\n` : ''}` +
      `## 页面内容预览\n\n` +
      `${content.length > 500 ? content.substring(0, 500) + '...' : content}`;

    return {
      content: [
        {
          type: 'text',
          text: resultText
        }
      ]
    };

  } catch (error) {
    console.error('创建页面失败:', error);
    
    if (axios.isAxiosError(error)) {
      const status = error.response?.status;
      const message = error.response?.data?.message || error.message;
      const errors = error.response?.data?.errors;
      
      if (status === 400) {
        let errorMessage = '创建页面失败，请检查参数:';
        if (errors && Array.isArray(errors)) {
          errorMessage += '\n' + errors.map((err: any) => `- ${err.message}`).join('\n');
        } else {
          errorMessage += ` ${message}`;
        }
        throw new Error(errorMessage);
      } else if (status === 403) {
        throw new Error(`没有权限在空间 ${spaceKey} 中创建页面`);
      } else if (status === 401) {
        throw new Error('认证失败，请检查API令牌和用户邮箱');
      } else if (status === 409) {
        throw new Error(`页面标题 "${title}" 在该空间中已存在`);
      } else {
        throw new Error(`创建页面失败: ${message}`);
      }
    }
    
    throw new Error(`创建页面时发生未知错误: ${error}`);
  }
}