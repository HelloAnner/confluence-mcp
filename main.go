package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// 验证必需的环境变量
	if os.Getenv("CONFLUENCE_BASE_URL") == "" ||
		os.Getenv("CONFLUENCE_USER_NAME") == "" ||
		os.Getenv("CONFLUENCE_API_TOKEN") == "" {
		log.Fatalf("❌ 缺少必需的环境变量:\n   CONFLUENCE_BASE_URL - Confluence 实例的基础 URL \n   CONFLUENCE_USER_NAME - Confluence 用户名称\n   CONFLUENCE_API_TOKEN - Confluence API 令牌")
	}

	// 创建 MCP 服务器
	s := server.NewMCPServer(
		"confluence-mcp",
		"1.0.0",
	)

	// 注册工具
	registerTools(s)

	// 创建 HTTP 服务器
	httpServer := server.NewStreamableHTTPServer(s)

	log.Println("Starting Confluence MCP HTTP server on :8080")
	log.Println("Endpoint: http://localhost:8080/mcp")
	log.Println("")
	log.Println("Available tools:")
	log.Println("- get_page: 获取Confluence页面信息")
	log.Println("- get_child_pages: 获取指定页面的子页面列表")
	log.Println("- create_page: 在Confluence中创建新页面")
	log.Println("- create_comment: 为Confluence页面添加评论")
	log.Println("- search_pages: 在Confluence中搜索页面")

	// 启动服务器
	if err := httpServer.Start(":8080"); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// 注册所有工具
func registerTools(s *server.MCPServer) {
	// 创建 Confluence 客户端
	client := NewConfluenceClient()

	log.Println(client.BaseURL);
	log.Println(client.Email);
	log.Println(client.APIToken);

	// 获取页面工具
	s.AddTool(mcp.NewTool("get_page",
		mcp.WithDescription("获取Confluence页面信息"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("Confluence页面的ID")),
	), handleGetPage(client))

	// 获取子页面工具
	s.AddTool(mcp.NewTool("get_child_pages",
		mcp.WithDescription("获取指定页面的子页面列表"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("父页面的ID")),
		mcp.WithNumber("limit", mcp.Description("返回结果的最大数量")),
		mcp.WithNumber("start", mcp.Description("起始位置")),
	), handleGetChildPages(client))

	// 创建页面工具
	s.AddTool(mcp.NewTool("create_page",
		mcp.WithDescription("在Confluence中创建新页面"),
		mcp.WithString("title", mcp.Required(), mcp.Description("页面标题")),
		mcp.WithString("content", mcp.Required(), mcp.Description("页面内容（支持Confluence存储格式）")),
		mcp.WithString("space_key", mcp.Required(), mcp.Description("空间键")),
		mcp.WithString("parent_id", mcp.Description("父页面ID（可选）")),
	), handleCreatePage(client))

	// 创建评论工具
	s.AddTool(mcp.NewTool("create_comment",
		mcp.WithDescription("为Confluence页面添加评论"),
		mcp.WithString("page_id", mcp.Required(), mcp.Description("页面ID")),
		mcp.WithString("comment", mcp.Required(), mcp.Description("评论内容")),
	), handleCreateComment(client))

	// 搜索页面工具
	s.AddTool(mcp.NewTool("search_pages",
		mcp.WithDescription("在Confluence中搜索页面"),
		mcp.WithString("query", mcp.Required(), mcp.Description("搜索关键词")),
		mcp.WithString("space_key", mcp.Description("限制搜索的空间（可选）")),
		mcp.WithNumber("limit", mcp.Description("返回结果的最大数量")),
		mcp.WithNumber("start", mcp.Description("起始位置")),
	), handleSearchPages(client))
}