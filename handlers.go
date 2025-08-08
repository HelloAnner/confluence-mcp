package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

var confluenceClient = NewConfluenceClient()

func handleGetPage(client *ConfluenceClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		pageID, err := request.RequireString("page_id")
		if err != nil {
			return mcp.NewToolResultError("page_id is required"), nil
		}

		page, err := client.GetPage(pageID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get page: %v", err)), nil
		}

		result, _ := json.Marshal(page)
		return mcp.NewToolResultText(string(result)), nil
	}
}

func handleGetChildPages(client *ConfluenceClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func handleCreatePage(client *ConfluenceClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		page, err := client.CreatePage(spaceKey, title, content, parentID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create page: %v", err)), nil
		}

		result, _ := json.Marshal(page)
		return mcp.NewToolResultText(string(result)), nil
	}
}

func handleCreateComment(client *ConfluenceClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func handleSearchPages(client *ConfluenceClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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