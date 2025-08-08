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

// 评论创建响应接口
interface CreateCommentResponse {
  id: string;
  title: string;
  type: string;
  status: string;
  body: {
    storage: {
      value: string;
      representation: string;
    };
  };
  version: {
    number: number;
    when: string;
    by: {
      displayName: string;
      email: string;
    };
  };
  container: {
    id: string;
    title: string;
  };
  _links: {
    webui: string;
  };
}

/**
 * 为指定页面创建评论
 * @param pageId 页面ID
 * @param comment 评论内容
 * @returns 创建的评论信息
 */
export async function createComment(pageId: string, comment: string) {
  try {
    // 验证页面是否存在
    let pageTitle = '';
    try {
      const pageResponse = await confluenceApi.get(`/content/${pageId}`);
      pageTitle = pageResponse.data.title;
    } catch (pageError) {
      if (axios.isAxiosError(pageError) && pageError.response?.status === 404) {
        throw new Error(`页面不存在 (ID: ${pageId})`);
      }
      throw pageError;
    }

    // 构建评论数据
    const commentData = {
      type: 'comment',
      container: {
        id: pageId,
        type: 'page'
      },
      body: {
        storage: {
          value: comment,
          representation: 'storage'
        }
      }
    };

    // 创建评论
    const response = await confluenceApi.post('/content', commentData);
    const createdComment: CreateCommentResponse = response.data;

    // 格式化返回结果
    const commentInfo = {
      id: createdComment.id,
      title: createdComment.title,
      type: createdComment.type,
      status: createdComment.status,
      content: createdComment.body.storage.value,
      contentFormat: createdComment.body.storage.representation,
      version: {
        number: createdComment.version.number,
        createdAt: createdComment.version.when,
        createdBy: {
          name: createdComment.version.by.displayName,
          email: createdComment.version.by.email
        }
      },
      parentPage: {
        id: createdComment.container.id,
        title: createdComment.container.title
      },
      webUrl: `${CONFLUENCE_BASE_URL}${createdComment._links.webui}`
    };

    const resultText = `# 评论创建成功\n\n` +
      `✅ 评论已成功添加到页面！\n\n` +
      `## 评论信息\n\n` +
      `- **评论ID:** ${commentInfo.id}\n` +
      `- **页面:** ${pageTitle} (ID: ${pageId})\n` +
      `- **状态:** ${commentInfo.status}\n` +
      `- **版本:** ${commentInfo.version.number}\n` +
      `- **创建时间:** ${new Date(commentInfo.version.createdAt).toLocaleString('zh-CN')}\n` +
      `- **创建者:** ${commentInfo.version.createdBy.name} (${commentInfo.version.createdBy.email})\n` +
      `- **评论链接:** [点击查看](${commentInfo.webUrl})\n\n` +
      `## 评论内容\n\n` +
      `${commentInfo.content}`;

    return {
      content: [
        {
          type: 'text',
          text: resultText
        }
      ]
    };

  } catch (error) {
    console.error('创建评论失败:', error);
    
    if (axios.isAxiosError(error)) {
      const status = error.response?.status;
      const message = error.response?.data?.message || error.message;
      const errors = error.response?.data?.errors;
      
      if (status === 400) {
        let errorMessage = '创建评论失败，请检查参数:';
        if (errors && Array.isArray(errors)) {
          errorMessage += '\n' + errors.map((err: any) => `- ${err.message}`).join('\n');
        } else {
          errorMessage += ` ${message}`;
        }
        throw new Error(errorMessage);
      } else if (status === 403) {
        throw new Error(`没有权限为页面 ${pageId} 创建评论`);
      } else if (status === 401) {
        throw new Error('认证失败，请检查API令牌和用户邮箱');
      } else if (status === 404) {
        throw new Error(`页面不存在 (ID: ${pageId})`);
      } else {
        throw new Error(`创建评论失败: ${message}`);
      }
    }
    
    throw new Error(`创建评论时发生未知错误: ${error}`);
  }
}