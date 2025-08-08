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

// 页面信息接口
interface PageInfo {
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
  body: {
    storage: {
      value: string;
      representation: string;
    };
  };
  _links: {
    webui: string;
  };
}

// 评论信息接口
interface CommentInfo {
  id: string;
  title: string;
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
}

/**
 * 根据页面ID获取Confluence页面内容和评论
 * @param pageId 页面ID
 * @param includeComments 是否包含评论
 * @returns 页面信息和评论列表
 */
export async function getPageById(pageId: string, includeComments: boolean = true) {
  try {
    // 获取页面基本信息
    const pageResponse = await confluenceApi.get(`/content/${pageId}`, {
      params: {
        expand: 'body.storage,version,space'
      }
    });

    const pageData: PageInfo = pageResponse.data;

    // 格式化页面信息
    const pageInfo = {
      id: pageData.id,
      title: pageData.title,
      type: pageData.type,
      status: pageData.status,
      space: {
        key: pageData.space.key,
        name: pageData.space.name
      },
      content: pageData.body.storage.value,
      contentFormat: pageData.body.storage.representation,
      version: {
        number: pageData.version.number,
        lastModified: pageData.version.when,
        lastModifiedBy: {
          name: pageData.version.by.displayName,
          email: pageData.version.by.email
        }
      },
      webUrl: `${CONFLUENCE_BASE_URL}${pageData._links.webui}`
    };

    let comments: any[] = [];

    // 如果需要获取评论
    if (includeComments) {
      try {
        const commentsResponse = await confluenceApi.get(`/content/${pageId}/child/comment`, {
          params: {
            expand: 'body.storage,version',
            limit: 100
          }
        });

        comments = commentsResponse.data.results.map((comment: CommentInfo) => ({
          id: comment.id,
          title: comment.title,
          content: comment.body.storage.value,
          contentFormat: comment.body.storage.representation,
          version: {
            number: comment.version.number,
            createdAt: comment.version.when,
            createdBy: {
              name: comment.version.by.displayName,
              email: comment.version.by.email
            }
          }
        }));
      } catch (commentError) {
        console.warn(`获取页面 ${pageId} 的评论时出错:`, commentError);
        // 即使获取评论失败，也返回页面信息
      }
    }

    return {
      content: [
        {
          type: 'text',
          text: `# 页面信息\n\n**标题:** ${pageInfo.title}\n**ID:** ${pageInfo.id}\n**空间:** ${pageInfo.space.name} (${pageInfo.space.key})\n**状态:** ${pageInfo.status}\n**版本:** ${pageInfo.version.number}\n**最后修改:** ${pageInfo.version.lastModified}\n**修改者:** ${pageInfo.version.lastModifiedBy.name}\n**链接:** ${pageInfo.webUrl}\n\n## 页面内容\n\n${pageInfo.content}\n\n${comments.length > 0 ? `## 评论 (${comments.length}条)\n\n${comments.map((comment, index) => `### 评论 ${index + 1}\n**创建者:** ${comment.version.createdBy.name}\n**创建时间:** ${comment.version.createdAt}\n\n${comment.content}\n`).join('\n')}` : '## 评论\n\n暂无评论'}`
        }
      ]
    };

  } catch (error) {
    console.error('获取页面信息失败:', error);
    
    if (axios.isAxiosError(error)) {
      const status = error.response?.status;
      const message = error.response?.data?.message || error.message;
      
      if (status === 404) {
        // 检查pageId是否有效
        if (/^\d+$/.test(pageId)) {
          throw new Error(`页面不存在但ID格式正确 (ID: ${pageId})，可能已被删除或移动`);
        } else {
          throw new Error(`无效的页面ID格式 (ID: ${pageId})`);
        }
      } else if (status === 403) {
        throw new Error(`没有权限访问页面 (ID: ${pageId})`);
      } else if (status === 401) {
        throw new Error('认证失败，请检查API令牌和用户邮箱');
      } else {
        throw new Error(`获取页面失败: ${message}`);
      }
    }
    
    throw new Error(`获取页面时发生未知错误: ${error}`);
  }
}