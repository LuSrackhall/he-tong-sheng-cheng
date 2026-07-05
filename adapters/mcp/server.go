package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// MCP Tool 定义
type MCPTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema struct {
		Type       string                        `json:"type"`
		Properties map[string]MCPToolProperty    `json:"properties"`
		Required   []string                      `json:"required"`
	} `json:"inputSchema"`
}

type MCPToolProperty struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// MCP 请求/响应
type MCPRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
	ID     int             `json:"id"`
}

type MCPResponse struct {
	ID     int         `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Knowledge Runtime 导入
// 通过 HTTP 调用后端 API 执行 capability
var apiBase = "http://localhost:8080"

func main() {
	if len(os.Args) > 1 {
		apiBase = os.Args[1]
	}

	http.HandleFunc("/mcp", handleMCP)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok","adapter":"mcp","version":"1.0.0"}`))
	})

	port := "9090"
	log.Printf("MCP Server starting on :%s (backend: %s)", port, apiBase)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(MCPResponse{
			ID: req.ID,
			Error: &MCPError{Code: -32700, Message: "Parse error: " + err.Error()},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch req.Method {
	case "initialize":
		json.NewEncoder(w).Encode(MCPResponse{
			ID:     req.ID,
			Result: map[string]any{"protocolVersion": "2025-03-26", "capabilities": map[string]any{"tools": map[string]any{}}},
		})

	case "tools/list":
		tools := []MCPTool{
			{
				Name:        "login",
				Description: "登录系统，获取认证令牌",
				InputSchema: struct {
					Type       string                        `json:"type"`
					Properties map[string]MCPToolProperty    `json:"properties"`
					Required   []string                      `json:"required"`
				}{
					Type: "object",
					Properties: map[string]MCPToolProperty{
						"username": {Type: "string", Description: "用户名"},
						"password": {Type: "string", Description: "密码"},
					},
					Required: []string{"username", "password"},
				},
			},
			{
				Name:        "collect_rent",
				Description: "录入租金收款并生成收据",
				InputSchema: struct {
					Type       string                        `json:"type"`
					Properties map[string]MCPToolProperty    `json:"properties"`
					Required   []string                      `json:"required"`
				}{
					Type: "object",
					Properties: map[string]MCPToolProperty{
						"contract_id": {Type: "number", Description: "合同 ID"},
						"amount":      {Type: "number", Description: "收款金额"},
						"date":        {Type: "string", Description: "收款日期 YYYY-MM-DD"},
					},
					Required: []string{"contract_id", "amount"},
				},
			},
			{
				Name:        "list_contracts",
				Description: "列出所有合同",
				InputSchema: struct {
					Type       string                        `json:"type"`
					Properties map[string]MCPToolProperty    `json:"properties"`
					Required   []string                      `json:"required"`
				}{
					Type:       "object",
					Properties: map[string]MCPToolProperty{},
				},
			},
			{
				Name:        "get_dashboard",
				Description: "获取仪表盘统计数据",
				InputSchema: struct {
					Type       string                        `json:"type"`
					Properties map[string]MCPToolProperty    `json:"properties"`
					Required   []string                      `json:"required"`
				}{
					Type:       "object",
					Properties: map[string]MCPToolProperty{},
				},
			},
			{
				Name:        "list_arrears",
				Description: "获取催缴清单",
				InputSchema: struct {
					Type       string                        `json:"type"`
					Properties map[string]MCPToolProperty    `json:"properties"`
					Required   []string                      `json:"required"`
				}{
					Type:       "object",
					Properties: map[string]MCPToolProperty{},
				},
			},
			{
				Name:        "create_contract",
				Description: "创建新合同",
				InputSchema: struct {
					Type       string                        `json:"type"`
					Properties map[string]MCPToolProperty    `json:"properties"`
					Required   []string                      `json:"required"`
				}{
					Type: "object",
					Properties: map[string]MCPToolProperty{
						"asset_id":    {Type: "number", Description: "资产 ID"},
						"tenant_id":   {Type: "number", Description: "租户 ID"},
						"start_date":  {Type: "string", Description: "开始日期 YYYY-MM-DD"},
						"end_date":    {Type: "string", Description: "结束日期 YYYY-MM-DD"},
						"monthly_rent": {Type: "number", Description: "月租金"},
					},
					Required: []string{"asset_id", "tenant_id", "start_date", "end_date", "monthly_rent"},
				},
			},
		}
		json.NewEncoder(w).Encode(MCPResponse{
			ID:     req.ID,
			Result: map[string]any{"tools": tools},
		})

	case "tools/call":
		var params struct {
			Name    string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		json.Unmarshal(req.Params, &params)
		result := executeCapability(params.Name, params.Arguments)
		json.NewEncoder(w).Encode(MCPResponse{
			ID:     req.ID,
			Result: result,
		})

	default:
		json.NewEncoder(w).Encode(MCPResponse{
			ID: req.ID,
			Error: &MCPError{Code: -32601, Message: fmt.Sprintf("Method %q not found", req.Method)},
		})
	}
}

func executeCapability(name string, args json.RawMessage) map[string]any {
	// 先通过 Kr Runtime 生成 Execution Plan
	// 然后调用后端 API 执行

	token := login()
	if token == "" {
		return map[string]any{"error": "authentication failed"}
	}

	switch name {
	case "login":
		return map[string]any{"message": "请直接通过 Kr CLI 登录"}

	case "collect_rent":
		var p struct {
			ContractID int     `json:"contract_id"`
			Amount     float64 `json:"amount"`
			Date       string  `json:"date"`
		}
		json.Unmarshal(args, &p)
		if p.Date == "" {
			p.Date = "2026-07-05"
		}
		resp, err := http.Post(
			fmt.Sprintf("%s/api/contracts/%d/payments", apiBase, p.ContractID),
			"application/json",
			strings.NewReader(fmt.Sprintf(`{"amount":%.2f,"date":"%s"}`, p.Amount, p.Date)),
		)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		defer resp.Body.Close()
		var result map[string]any
		json.NewDecoder(resp.Body).Decode(&result)
		result["_status"] = resp.StatusCode
		return result

	case "list_contracts":
		resp, err := httpGet(fmt.Sprintf("%s/api/contracts", apiBase), token)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"contracts": resp}

	case "get_dashboard":
		resp, err := httpGet(fmt.Sprintf("%s/api/dashboard/stats", apiBase), token)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return resp

	case "list_arrears":
		resp, err := httpGet(fmt.Sprintf("%s/api/arrears", apiBase), token)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return map[string]any{"arrears": resp}

	default:
		return map[string]any{"error": fmt.Sprintf("unknown capability: %s", name)}
	}
}

func login() string {
	resp, err := http.Post(
		fmt.Sprintf("%s/api/auth/login", apiBase),
		"application/json",
		strings.NewReader(`{"username":"admin","password":"admin123"}`),
	)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	token, _ := result["token"].(string)
	return token
}

func httpGet(url, token string) (map[string]any, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	result["_status"] = resp.StatusCode
	return result, nil
}
