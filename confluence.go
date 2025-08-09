package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ConfluenceClient Confluence API 客户端
type ConfluenceClient struct {
	BaseURL    string
	Email      string
	APIToken   string
	HTTPClient *http.Client
}

// NewConfluenceClient 创建新的 Confluence 客户端
func NewConfluenceClient() *ConfluenceClient {
	return &ConfluenceClient{
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewConfluenceClientWithCredentials 使用指定凭据创建客户端
func NewConfluenceClientWithCredentials(baseURL, email, apiToken string) *ConfluenceClient {
	return &ConfluenceClient{
		BaseURL:  baseURL,
		Email:    email,
		APIToken: apiToken,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetCredentials 设置客户端凭据
func (c *ConfluenceClient) SetCredentials(baseURL, email, apiToken string) {
	c.BaseURL = baseURL
	c.Email = email
	c.APIToken = apiToken
}

// ValidateCredentials 验证凭据是否完整
func (c *ConfluenceClient) ValidateCredentials() error {
	if c.BaseURL == "" {
		return fmt.Errorf("缺少 Confluence Base URL")
	}
	if c.Email == "" {
		return fmt.Errorf("缺少用户邮箱")
	}
	if c.APIToken == "" {
		return fmt.Errorf("缺少 API Token")
	}
	return nil
}

// makeRequest 发送 HTTP 请求
func (c *ConfluenceClient) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := fmt.Sprintf("%s/rest/api%s", c.BaseURL, endpoint)
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.SetBasicAuth(c.Email, c.APIToken)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 请求失败 (状态码: %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

// PageInfo 页面信息结构
type PageInfo struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Space  struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	} `json:"space"`
	Version struct {
		Number int    `json:"number"`
		When   string `json:"when"`
		By     struct {
			DisplayName string `json:"displayName"`
			Email       string `json:"email"`
		} `json:"by"`
	} `json:"version"`
	Body struct {
		Storage struct {
			Value          string `json:"value"`
			Representation string `json:"representation"`
		} `json:"storage"`
	} `json:"body"`
	Links struct {
		WebUI string `json:"webui"`
	} `json:"_links"`
}

// CommentInfo 评论信息结构
type CommentInfo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
	Body  struct {
		Storage struct {
			Value          string `json:"value"`
			Representation string `json:"representation"`
		} `json:"storage"`
	} `json:"body"`
	Version struct {
		Number int    `json:"number"`
		When   string `json:"when"`
		By     struct {
			DisplayName string `json:"displayName"`
			Email       string `json:"email"`
		} `json:"by"`
	} `json:"version"`
	Container struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"container"`
	Links struct {
		WebUI string `json:"webui"`
	} `json:"_links"`
}

// ChildPageInfo 子页面信息结构
type ChildPageInfo struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Type    string `json:"type"`
	Status  string `json:"status"`
	Version struct {
		Number int    `json:"number"`
		When   string `json:"when"`
		By     struct {
			DisplayName string `json:"displayName"`
			Email       string `json:"email"`
		} `json:"by"`
	} `json:"version"`
	History struct {
		CreatedDate string `json:"createdDate"`
		CreatedBy   struct {
			DisplayName string `json:"displayName"`
			Email       string `json:"email"`
		} `json:"createdBy"`
	} `json:"history"`
	Links struct {
		WebUI string `json:"webui"`
	} `json:"_links"`
}

// SearchResult 搜索结果结构
type SearchResult struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Space  struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	} `json:"space"`
	Version struct {
		Number int    `json:"number"`
		When   string `json:"when"`
		By     struct {
			DisplayName string `json:"displayName"`
			Email       string `json:"email"`
		} `json:"by"`
	} `json:"version"`
	History struct {
		CreatedDate string `json:"createdDate"`
		CreatedBy   struct {
			DisplayName string `json:"displayName"`
			Email       string `json:"email"`
		} `json:"createdBy"`
	} `json:"history"`
	Body *struct {
		Storage struct {
			Value          string `json:"value"`
			Representation string `json:"representation"`
		} `json:"storage"`
	} `json:"body,omitempty"`
	Excerpt *string `json:"excerpt,omitempty"`
	Links   struct {
		WebUI string `json:"webui"`
	} `json:"_links"`
}

// GetPageComments 获取页面评论
func (c *ConfluenceClient) GetPageComments(pageID string) ([]CommentInfo, error) {
	endpoint := fmt.Sprintf("/content/%s/child/comment?expand=body.storage,version", pageID)
	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("获取页面评论失败: %w", err)
	}
	defer resp.Body.Close()

	var commentsResponse struct {
		Results []CommentInfo `json:"results"`
		Start   int           `json:"start"`
		Limit   int           `json:"limit"`
		Size    int           `json:"size"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&commentsResponse); err != nil {
		return nil, fmt.Errorf("解析评论数据失败: %w", err)
	}

	return commentsResponse.Results, nil
}

// GetPage 获取页面信息（包含评论）
func (c *ConfluenceClient) GetPage(pageID string) (*PageWithCommentsResponse, error) {
	// 获取页面信息
	resp, err := c.makeRequest("GET", fmt.Sprintf("/content/%s?expand=body.storage,version,space", pageID), nil)
	if err != nil {
		return nil, fmt.Errorf("获取页面失败: %w", err)
	}
	defer resp.Body.Close()

	var page PageResponse
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("解析页面数据失败: %w", err)
	}

	// 获取页面评论
	comments, err := c.GetPageComments(pageID)
	if err != nil {
		// 如果获取评论失败，记录错误但不影响页面获取
		comments = []CommentInfo{}
	}

	// 组合页面和评论数据
	result := &PageWithCommentsResponse{
		Page:     page,
		Comments: comments,
	}

	return result, nil
}

// GetChildPages 获取子页面列表
func (c *ConfluenceClient) GetChildPages(pageID string, limit, start int) (*ChildPagesResponse, error) {
	endpoint := fmt.Sprintf("/content/%s/child/page?expand=version,history&limit=%d&start=%d", pageID, limit, start)
	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("获取子页面失败: %w", err)
	}
	defer resp.Body.Close()

	var childPages ChildPagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&childPages); err != nil {
		return nil, fmt.Errorf("解析子页面数据失败: %w", err)
	}

	return &childPages, nil
}

// CreatePageRequest 创建页面请求结构
type CreatePageRequest struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Space struct {
		Key string `json:"key"`
	} `json:"space"`
	Body struct {
		Storage struct {
			Value          string `json:"value"`
			Representation string `json:"representation"`
		} `json:"storage"`
	} `json:"body"`
	Ancestors []struct {
		ID string `json:"id"`
	} `json:"ancestors,omitempty"`
}

// Response types
type PageResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
	Space struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	} `json:"space"`
	Body struct {
		Storage struct {
			Value          string `json:"value"`
			Representation string `json:"representation"`
		} `json:"storage"`
	} `json:"body"`
	Version struct {
		Number int    `json:"number"`
		When   string `json:"when"`
		By     struct {
			DisplayName string `json:"displayName"`
			Email       string `json:"email"`
		} `json:"by"`
	} `json:"version"`
	Links struct {
		Webui string `json:"webui"`
	} `json:"_links"`
}

type PageWithCommentsResponse struct {
	Page     PageResponse  `json:"page"`
	Comments []CommentInfo `json:"comments"`
}

type ChildPagesResponse struct {
	Results []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Type  string `json:"type"`
		Links struct {
			Webui string `json:"webui"`
		} `json:"_links"`
	} `json:"results"`
	Start int `json:"start"`
	Limit int `json:"limit"`
	Size  int `json:"size"`
}

type SearchResponse struct {
	Results []struct {
		ID      string `json:"id"`
		Title   string `json:"title"`
		Type    string `json:"type"`
		Excerpt string `json:"excerpt"`
		Space   struct {
			Key  string `json:"key"`
			Name string `json:"name"`
		} `json:"space"`
		Links struct {
			Webui string `json:"webui"`
		} `json:"_links"`
	} `json:"results"`
	Start int `json:"start"`
	Limit int `json:"limit"`
	Size  int `json:"size"`
}

// CreatePage 创建页面
func (c *ConfluenceClient) CreatePage(title, content, spaceKey, parentID string) (*PageResponse, error) {
	req := CreatePageRequest{
		Type:  "page",
		Title: title,
	}
	req.Space.Key = spaceKey
	req.Body.Storage.Value = content
	req.Body.Storage.Representation = "storage"

	if parentID != "" {
		req.Ancestors = []struct {
			ID string `json:"id"`
		}{{ID: parentID}}
	}

	resp, err := c.makeRequest("POST", "/content", req)
	if err != nil {
		return nil, fmt.Errorf("创建页面失败: %w", err)
	}
	defer resp.Body.Close()

	var page PageResponse
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("解析创建页面响应失败: %w", err)
	}

	return &page, nil
}

// CreateCommentRequest 创建评论请求结构
type CreateCommentRequest struct {
	Type      string `json:"type"`
	Container struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	} `json:"container"`
	Body struct {
		Storage struct {
			Value          string `json:"value"`
			Representation string `json:"representation"`
		} `json:"storage"`
	} `json:"body"`
}

// CreateComment 创建评论
func (c *ConfluenceClient) CreateComment(pageID, comment string) (*CommentInfo, error) {
	req := CreateCommentRequest{
		Type: "comment",
	}
	req.Container.ID = pageID
	req.Container.Type = "page"
	req.Body.Storage.Value = comment
	req.Body.Storage.Representation = "storage"

	resp, err := c.makeRequest("POST", "/content", req)
	if err != nil {
		return nil, fmt.Errorf("创建评论失败: %w", err)
	}
	defer resp.Body.Close()

	var commentInfo CommentInfo
	if err := json.NewDecoder(resp.Body).Decode(&commentInfo); err != nil {
		return nil, fmt.Errorf("解析创建评论响应失败: %w", err)
	}

	return &commentInfo, nil
}

// SearchPages 搜索页面
func (c *ConfluenceClient) SearchPages(query, spaceKey string, limit, start int) (*SearchResponse, error) {
	params := url.Values{}
	params.Set("cql", fmt.Sprintf("text ~ \"%s\" and type = page", query))
	if spaceKey != "" {
		params.Set("cql", fmt.Sprintf("text ~ \"%s\" and type = page and space = %s", query, spaceKey))
	}
	params.Set("limit", strconv.Itoa(limit))
	params.Set("start", strconv.Itoa(start))
	params.Set("expand", "version,history,body.storage")

	endpoint := fmt.Sprintf("/content/search?%s", params.Encode())
	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("搜索页面失败: %w", err)
	}
	defer resp.Body.Close()

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("解析搜索结果失败: %w", err)
	}

	return &searchResp, nil
}

// MarkdownPageResponse Markdown格式的页面响应
type MarkdownPageResponse struct {
	Metadata MarkdownMetadata `json:"metadata"`
	Content  string           `json:"content"`
}

// MarkdownMetadata 页面元数据
type MarkdownMetadata struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	SpaceKey    string    `json:"space_key"`
	SpaceName   string    `json:"space_name"`
	Version     int       `json:"version"`
	LastUpdated time.Time `json:"last_updated"`
	UpdatedBy   string    `json:"updated_by"`
	WebURL      string    `json:"web_url"`
}

// ConvertPageToMarkdown 将页面内容转换为Markdown格式
func (c *ConfluenceClient) ConvertPageToMarkdown(pageID string) (*MarkdownPageResponse, error) {
	// 获取页面和评论数据
	pageWithComments, err := c.GetPage(pageID)
	if err != nil {
		return nil, fmt.Errorf("获取页面数据失败: %w", err)
	}

	// 解析最后更新时间
	lastUpdated, _ := time.Parse(time.RFC3339, pageWithComments.Page.Version.When)

	// 构建元数据
	metadata := MarkdownMetadata{
		ID:          pageWithComments.Page.ID,
		Title:       pageWithComments.Page.Title,
		SpaceKey:    pageWithComments.Page.Space.Key,
		SpaceName:   pageWithComments.Page.Space.Name,
		Version:     pageWithComments.Page.Version.Number,
		LastUpdated: lastUpdated,
		UpdatedBy:   pageWithComments.Page.Version.By.DisplayName,
		WebURL:      c.BaseURL + pageWithComments.Page.Links.Webui,
	}

	// 转换页面内容为Markdown
	markdownContent := c.convertToMarkdown(pageWithComments)

	return &MarkdownPageResponse{
		Metadata: metadata,
		Content:  markdownContent,
	}, nil
}

// convertToMarkdown 将Confluence存储格式转换为Markdown
func (c *ConfluenceClient) convertToMarkdown(pageWithComments *PageWithCommentsResponse) string {
	var markdown strings.Builder

	// 添加页面标题和元数据
	markdown.WriteString(fmt.Sprintf("# %s\n\n", pageWithComments.Page.Title))

	// 添加元数据信息
	markdown.WriteString("## 页面信息\n\n")
	markdown.WriteString(fmt.Sprintf("- **页面ID**: %s\n", pageWithComments.Page.ID))
	markdown.WriteString(fmt.Sprintf("- **空间**: %s (%s)\n", pageWithComments.Page.Space.Name, pageWithComments.Page.Space.Key))
	markdown.WriteString(fmt.Sprintf("- **版本**: %d\n", pageWithComments.Page.Version.Number))
	if pageWithComments.Page.Version.When != "" {
		if lastUpdated, err := time.Parse(time.RFC3339, pageWithComments.Page.Version.When); err == nil {
			markdown.WriteString(fmt.Sprintf("- **最后更新**: %s\n", lastUpdated.Format("2006-01-02 15:04:05")))
		}
	}
	if pageWithComments.Page.Version.By.DisplayName != "" {
		markdown.WriteString(fmt.Sprintf("- **更新者**: %s\n", pageWithComments.Page.Version.By.DisplayName))
	}
	markdown.WriteString(fmt.Sprintf("- **页面链接**: %s%s\n\n", c.BaseURL, pageWithComments.Page.Links.Webui))

	// 添加页面内容
	markdown.WriteString("## 页面内容\n\n")
	pageContent := c.htmlToMarkdown(pageWithComments.Page.Body.Storage.Value)
	markdown.WriteString(pageContent)
	markdown.WriteString("\n\n")

	// 添加评论部分
	if len(pageWithComments.Comments) > 0 {
		markdown.WriteString("## 评论\n\n")
		for i, comment := range pageWithComments.Comments {
			markdown.WriteString(fmt.Sprintf("### 评论 %d\n\n", i+1))

			// 评论元数据
			if comment.Version.By.DisplayName != "" {
				markdown.WriteString(fmt.Sprintf("**作者**: %s  \n", comment.Version.By.DisplayName))
			}
			if comment.Version.When != "" {
				if commentTime, err := time.Parse(time.RFC3339, comment.Version.When); err == nil {
					markdown.WriteString(fmt.Sprintf("**时间**: %s  \n", commentTime.Format("2006-01-02 15:04:05")))
				}
			}
			markdown.WriteString("\n")

			// 评论内容
			commentContent := c.htmlToMarkdown(comment.Body.Storage.Value)
			markdown.WriteString(commentContent)
			markdown.WriteString("\n\n")
		}
	}

	return markdown.String()
}

// htmlToMarkdown 将Confluence HTML存储格式转换为Markdown
func (c *ConfluenceClient) htmlToMarkdown(html string) string {
	// 移除XML声明和根标签
	content := strings.TrimSpace(html)

	// 移除常见的XML声明
	if strings.HasPrefix(content, "<?xml") {
		if idx := strings.Index(content, "?>"); idx != -1 {
			content = content[idx+2:]
		}
	}

	// 基本的HTML到Markdown转换
	content = c.convertBasicHTML(content)
	content = c.convertConfluenceSpecific(content)
	content = c.cleanupMarkdown(content)

	return strings.TrimSpace(content)
}

// convertBasicHTML 转换基本的HTML标签
func (c *ConfluenceClient) convertBasicHTML(content string) string {
	// 标题转换
	content = regexp.MustCompile(`<h([1-6])(?:[^>]*)>(.*?)</h[1-6]>`).ReplaceAllStringFunc(content, func(match string) string {
		re := regexp.MustCompile(`<h([1-6])(?:[^>]*)>(.*?)</h[1-6]>`)
		matches := re.FindStringSubmatch(match)
		if len(matches) >= 3 {
			level := matches[1]
			text := strings.TrimSpace(matches[2])
			text = c.stripHTMLTags(text)
			return strings.Repeat("#", parseInt(level)) + " " + text + "\n\n"
		}
		return match
	})

	// 段落转换
	content = regexp.MustCompile(`<p(?:[^>]*)>(.*?)</p>`).ReplaceAllStringFunc(content, func(match string) string {
		re := regexp.MustCompile(`<p(?:[^>]*)>(.*?)</p>`)
		matches := re.FindStringSubmatch(match)
		if len(matches) >= 2 {
			text := strings.TrimSpace(matches[1])
			if text != "" {
				return text + "\n\n"
			}
		}
		return ""
	})

	// 粗体转换
	content = regexp.MustCompile(`<strong(?:[^>]*)>(.*?)</strong>`).ReplaceAllString(content, "**$1**")
	content = regexp.MustCompile(`<b(?:[^>]*)>(.*?)</b>`).ReplaceAllString(content, "**$1**")

	// 斜体转换
	content = regexp.MustCompile(`<em(?:[^>]*)>(.*?)</em>`).ReplaceAllString(content, "*$1*")
	content = regexp.MustCompile(`<i(?:[^>]*)>(.*?)</i>`).ReplaceAllString(content, "*$1*")

	// 代码转换
	content = regexp.MustCompile(`<code(?:[^>]*)>(.*?)</code>`).ReplaceAllString(content, "`$1`")

	// 链接转换
	content = regexp.MustCompile(`<a(?:[^>]*)\s+href="([^"]*)"(?:[^>]*)>(.*?)</a>`).ReplaceAllString(content, "[$2]($1)")

	// 换行转换
	content = regexp.MustCompile(`<br\s*/?>`).ReplaceAllString(content, "  \n")

	// 列表转换
	content = c.convertLists(content)

	return content
}

// convertLists 转换列表
func (c *ConfluenceClient) convertLists(content string) string {
	// 无序列表
	content = regexp.MustCompile(`<ul(?:[^>]*)>(.*?)</ul>`).ReplaceAllStringFunc(content, func(match string) string {
		re := regexp.MustCompile(`<ul(?:[^>]*)>(.*?)</ul>`)
		matches := re.FindStringSubmatch(match)
		if len(matches) >= 2 {
			listContent := matches[1]
			listContent = regexp.MustCompile(`<li(?:[^>]*)>(.*?)</li>`).ReplaceAllString(listContent, "- $1\n")
			return "\n" + listContent + "\n"
		}
		return match
	})

	// 有序列表
	content = regexp.MustCompile(`<ol(?:[^>]*)>(.*?)</ol>`).ReplaceAllStringFunc(content, func(match string) string {
		re := regexp.MustCompile(`<ol(?:[^>]*)>(.*?)</ol>`)
		matches := re.FindStringSubmatch(match)
		if len(matches) >= 2 {
			listContent := matches[1]
			counter := 1
			listContent = regexp.MustCompile(`<li(?:[^>]*)>(.*?)</li>`).ReplaceAllStringFunc(listContent, func(item string) string {
				re := regexp.MustCompile(`<li(?:[^>]*)>(.*?)</li>`)
				itemMatches := re.FindStringSubmatch(item)
				if len(itemMatches) >= 2 {
					result := fmt.Sprintf("%d. %s\n", counter, itemMatches[1])
					counter++
					return result
				}
				return item
			})
			return "\n" + listContent + "\n"
		}
		return match
	})

	return content
}

// convertConfluenceSpecific 转换Confluence特定的标签
func (c *ConfluenceClient) convertConfluenceSpecific(content string) string {
	// Confluence代码块
	content = regexp.MustCompile(`<ac:structured-macro[^>]*ac:name="code"[^>]*>(.*?)</ac:structured-macro>`).ReplaceAllStringFunc(content, func(match string) string {
		// 提取语言参数
		langRe := regexp.MustCompile(`<ac:parameter[^>]*ac:name="language"[^>]*>(.*?)</ac:parameter>`)
		langMatches := langRe.FindStringSubmatch(match)
		language := ""
		if len(langMatches) >= 2 {
			language = langMatches[1]
		}

		// 提取代码内容
		codeRe := regexp.MustCompile(`<ac:plain-text-body><!\[CDATA\[(.*?)\]\]></ac:plain-text-body>`)
		codeMatches := codeRe.FindStringSubmatch(match)
		if len(codeMatches) >= 2 {
			code := codeMatches[1]
			return fmt.Sprintf("\n```%s\n%s\n```\n\n", language, code)
		}
		return match
	})

	// Confluence信息框
	content = regexp.MustCompile(`<ac:structured-macro[^>]*ac:name="info"[^>]*>(.*?)</ac:structured-macro>`).ReplaceAllStringFunc(content, func(match string) string {
		bodyRe := regexp.MustCompile(`<ac:rich-text-body>(.*?)</ac:rich-text-body>`)
		bodyMatches := bodyRe.FindStringSubmatch(match)
		if len(bodyMatches) >= 2 {
			body := c.stripHTMLTags(bodyMatches[1])
			return fmt.Sprintf("\n> ℹ️ **信息**: %s\n\n", body)
		}
		return match
	})

	// Confluence警告框
	content = regexp.MustCompile(`<ac:structured-macro[^>]*ac:name="warning"[^>]*>(.*?)</ac:structured-macro>`).ReplaceAllStringFunc(content, func(match string) string {
		bodyRe := regexp.MustCompile(`<ac:rich-text-body>(.*?)</ac:rich-text-body>`)
		bodyMatches := bodyRe.FindStringSubmatch(match)
		if len(bodyMatches) >= 2 {
			body := c.stripHTMLTags(bodyMatches[1])
			return fmt.Sprintf("\n> ⚠️ **警告**: %s\n\n", body)
		}
		return match
	})

	// Confluence表格 (简化处理)
	content = regexp.MustCompile(`<table[^>]*>(.*?)</table>`).ReplaceAllStringFunc(content, func(match string) string {
		// 简单的表格转换，实际实现可能需要更复杂的逻辑
		return "\n[表格内容 - 需要手动格式化]\n\n"
	})

	return content
}

// stripHTMLTags 移除HTML标签
func (c *ConfluenceClient) stripHTMLTags(content string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(content, "")
}

// cleanupMarkdown 清理Markdown格式
func (c *ConfluenceClient) cleanupMarkdown(content string) string {
	// 移除多余的空行
	content = regexp.MustCompile(`\n{3,}`).ReplaceAllString(content, "\n\n")

	// 移除行首行尾的空格
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	content = strings.Join(lines, "\n")

	return content
}

// parseInt 辅助函数：字符串转整数
func parseInt(s string) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return 1
}
