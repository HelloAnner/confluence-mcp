package main

import (
	"context"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// 加载环境变量（可选，用于默认配置）
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// 创建 MCP 服务器
	s := server.NewMCPServer(
		"confluence-mcp",
		"1.0.0",
	)

	// 注册工具
	registerTools(s)

	// 创建 HTTP 服务器，配置HTTP上下文函数来提取头信息
	httpServer := server.NewStreamableHTTPServer(s,
		server.WithHTTPContextFunc(func(ctx context.Context, req *http.Request) context.Context {
			return req.Context() // Do Nothing
		}),
	)

	log.Println("Starting Confluence MCP HTTP server on :8080")
	log.Println("Endpoint: http://localhost:8080/mcp")
	log.Println("")
	log.Println("Multi-user support enabled - pass credentials via headers:")
	log.Println("- X-Confluence-Base-URL: Confluence Address")
	log.Println("- X-Confluence-Name: UserName")
	log.Println("- X-Confluence-Token: UserPassword")
	log.Println("")
	log.Println("Available tools:")
	log.Println("- get_page: 获取Confluence页面并返回Markdown格式（包含页面内容和评论）")
	log.Println("- get_child_pages: 获取指定页面的子页面列表")
	log.Println("- create_page: 在Confluence中创建新页面")
	log.Println("- create_comment: 为Confluence页面添加评论")
	log.Println("- search_pages: 在Confluence中搜索页面")
	log.Println("- convert_page_to_markdown: 将Confluence页面转换为Markdown格式（返回JSON格式的元数据）")

	// 启动服务器
	if err := httpServer.Start(":8080"); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// 注册所有工具
func registerTools(s *server.MCPServer) {
	// 获取页面工具
	s.AddTool(mcp.NewTool("get_page_and_comment",
		mcp.WithDescription("获取Confluence页面内容并返回Markdown格式（包含页面元数据、内容和评论）"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("Confluence页面的ID")),
	), handleGetPage())

	// 获取子页面工具
	s.AddTool(mcp.NewTool("get_child_page_list",
		mcp.WithDescription("获取指定页面的子页面列表"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("父页面的ID")),
		mcp.WithNumber("limit", mcp.Description("返回结果的最大数量")),
		mcp.WithNumber("start", mcp.Description("起始位置")),
	), handleGetChildPages())

	// 创建页面工具
	s.AddTool(mcp.NewTool("create_new_page",
		mcp.WithDescription("在Confluence中创建新页面"),
		mcp.WithString("title", mcp.Required(), mcp.Description("页面标题")),
		mcp.WithString("content", mcp.Required(), mcp.Description("页面内容（支持Confluence存储格式）")),
		mcp.WithString("space_key", mcp.Required(), mcp.Description("空间键")),
		mcp.WithString("parent_id", mcp.Description("父页面ID（可选）")),
	), handleCreatePage())

	// 创建评论工具
	s.AddTool(mcp.NewTool("create_new_comment",
		mcp.WithDescription("为Confluence页面添加评论"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("页面ID")),
		mcp.WithString("comment", mcp.Required(), mcp.Description("评论内容")),
	), handleCreateComment())

	// 搜索页面工具
	s.AddTool(mcp.NewTool("search_pages",
		mcp.WithDescription("在Confluence中搜索页面"),
		mcp.WithString("query", mcp.Required(), mcp.Description("搜索关键词")),
		mcp.WithString("space_key", mcp.Description("限制搜索的空间（可选）")),
		mcp.WithNumber("limit", mcp.Description("返回结果的最大数量")),
		mcp.WithNumber("start", mcp.Description("起始位置")),
	), handleSearchPages())

	// 转换页面为Markdown工具
	s.AddTool(mcp.NewTool("convert_page_to_markdown",
		mcp.WithDescription("将Confluence页面内容转换为Markdown格式，包含页面元数据、内容和评论"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("要转换的Confluence页面ID")),
	), handleConvertPageToMarkdown())
}
