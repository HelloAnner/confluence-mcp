package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func handleGetPage() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := getClientFromContext(request)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("认证失败: %v", err)), nil
		}

		pageID, err := request.RequireString("page_id")
		if err != nil {
			return mcp.NewToolResultError("page_id is required"), nil
		}

		// 直接转换为Markdown格式
		markdownPage, err := client.ConvertPageToMarkdown(pageID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get page as markdown: %v", err)), nil
		}

		// 返回纯Markdown文本内容
		return mcp.NewToolResultText(markdownPage.Content), nil
	}
}

func handleGetChildPages() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := getClientFromContext(request)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("认证失败: %v", err)), nil
		}

		pageID, err := request.RequireString("page_id")
		if err != nil {
			return mcp.NewToolResultError("page_id is required"), nil
		}

		limit := request.GetInt("limit", 25)
		start := request.GetInt("start", 0)

		children, err := client.GetChildPages(pageID, limit, start)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get child pages: %v", err)), nil
		}

		result, _ := json.Marshal(children)
		return mcp.NewToolResultText(string(result)), nil
	}
}

func handleCreatePage() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := getClientFromContext(request)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("认证失败: %v", err)), nil
		}

		title, err := request.RequireString("title")
		if err != nil {
			return mcp.NewToolResultError("title is required"), nil
		}

		content, err := request.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError("content is required"), nil
		}

		spaceKey, err := request.RequireString("space_key")
		if err != nil {
			return mcp.NewToolResultError("space_key is required"), nil
		}

		parentID := request.GetString("parent_id", "")

		page, err := client.CreatePage(title, content, spaceKey, parentID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create page: %v", err)), nil
		}

		result, _ := json.Marshal(page)
		return mcp.NewToolResultText(string(result)), nil
	}
}

func handleCreateComment() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := getClientFromContext(request)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("认证失败: %v", err)), nil
		}

		pageID, err := request.RequireString("page_id")
		if err != nil {
			return mcp.NewToolResultError("page_id is required"), nil
		}

		comment, err := request.RequireString("comment")
		if err != nil {
			return mcp.NewToolResultError("comment is required"), nil
		}

		result, err := client.CreateComment(pageID, comment)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create comment: %v", err)), nil
		}

		response, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(response)), nil
	}
}

func handleSearchPages() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := getClientFromContext(request)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("认证失败: %v", err)), nil
		}

		query, err := request.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError("query is required"), nil
		}

		limit := request.GetInt("limit", 25)
		spaceKey := request.GetString("space_key", "")
		start := request.GetInt("start", 0)

		pages, err := client.SearchPages(query, spaceKey, limit, start)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to search pages: %v", err)), nil
		}

		result, _ := json.Marshal(pages)
		return mcp.NewToolResultText(string(result)), nil
	}
}

func handleConvertPageToMarkdown() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		client, err := getClientFromContext(request)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("认证失败: %v", err)), nil
		}

		pageID, err := request.RequireString("page_id")
		if err != nil {
			return mcp.NewToolResultError("page_id is required"), nil
		}

		markdownPage, err := client.ConvertPageToMarkdown(pageID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to convert page to markdown: %v", err)), nil
		}

		result, _ := json.Marshal(markdownPage)
		return mcp.NewToolResultText(string(result)), nil
	}
}

// getClientFromContext 从上下文中获取用户凭据并创建客户端
func getClientFromContext(request mcp.CallToolRequest) (*ConfluenceClient, error) {

	headers := request.Header

	// 首先尝试从环境变量获取默认配置
	baseURL := headers.Get("X-Confluence-Base-URL")
	name := headers.Get("X-Confluence-Name")
	apiToken := headers.Get("X-Confluence-Token")

	client := NewConfluenceClientWithCredentials(baseURL, name, apiToken)
	if err := client.ValidateCredentials(); err != nil {
		return nil, fmt.Errorf("confluence auth failed: %v", err)
	}

	return client, nil
}
