package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// ConfluenceClient Confluence API 客户端
type ConfluenceClient struct {
	BaseURL   string
	Email     string
	APIToken  string
	HTTPClient *http.Client
}

// NewConfluenceClient 创建新的 Confluence 客户端
func NewConfluenceClient() *ConfluenceClient {
	return &ConfluenceClient{
		BaseURL:  os.Getenv("CONFLUENCE_BASE_URL"),
		Email:    os.Getenv("CONFLUENCE_USER_EMAIL"),
		APIToken: os.Getenv("CONFLUENCE_API_TOKEN"),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
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
	ID     string `json:"id"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	Status string `json:"status"`
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

// GetPage 获取页面信息
func (c *ConfluenceClient) GetPage(pageID string) (*PageResponse, error) {
	resp, err := c.makeRequest("GET", fmt.Sprintf("/content/%s?expand=body.storage,version,space", pageID), nil)
	if err != nil {
		return nil, fmt.Errorf("获取页面失败: %w", err)
	}
	defer resp.Body.Close()

	var page PageResponse
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("解析页面数据失败: %w", err)
	}

	return &page, nil
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
		Number int `json:"number"`
	} `json:"version"`
	Links struct {
		Webui string `json:"webui"`
	} `json:"_links"`
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