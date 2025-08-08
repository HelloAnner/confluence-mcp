import { Request, Response, NextFunction } from 'express';

// 用户配置接口
export interface UserConfig {
  confluenceBaseUrl: string;
  confluenceApiToken: string;
  confluenceUserEmail: string;
}

// 从请求中提取用户配置
export function extractUserConfig(req: Request): UserConfig | null {
  const confluenceBaseUrl = req.headers['x-confluence-base-url'] as string;
  const confluenceApiToken = req.headers['x-confluence-api-token'] as string;
  const confluenceUserEmail = req.headers['x-confluence-user-email'] as string;

  // 如果请求头中有完整的配置，使用请求头的配置
  if (confluenceBaseUrl && confluenceApiToken && confluenceUserEmail) {
    return {
      confluenceBaseUrl,
      confluenceApiToken,
      confluenceUserEmail
    };
  }

  // 否则尝试使用环境变量
  if (process.env.CONFLUENCE_BASE_URL && process.env.CONFLUENCE_API_TOKEN && process.env.CONFLUENCE_USER_EMAIL) {
    return {
      confluenceBaseUrl: process.env.CONFLUENCE_BASE_URL,
      confluenceApiToken: process.env.CONFLUENCE_API_TOKEN,
      confluenceUserEmail: process.env.CONFLUENCE_USER_EMAIL
    };
  }

  return null;
}

// 认证中间件
export function authMiddleware(req: Request, res: Response, next: NextFunction) {
  const userConfig = extractUserConfig(req);
  
  if (!userConfig) {
    return res.status(401).json({
      error: 'Authentication required',
      message: '请在请求头中提供 Confluence 认证信息：X-Confluence-Base-URL, X-Confluence-API-Token, X-Confluence-User-Email'
    });
  }

  // 将用户配置附加到请求对象
  (req as any).userConfig = userConfig;
  next();
}

// 可选认证中间件（允许使用环境变量作为后备）
export function optionalAuthMiddleware(req: Request, res: Response, next: NextFunction) {
  const userConfig = extractUserConfig(req);
  
  if (userConfig) {
    (req as any).userConfig = userConfig;
  }
  
  next();
}