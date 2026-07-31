package qoder

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SessionInfo 会话元数据
type SessionInfo struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	WorkingDir   string `json:"workingDir"`
	CreatedAt    int64  `json:"createdAt"`
	UpdatedAt    int64  `json:"updatedAt"`
	MessageCount int    `json:"messageCount"`
}

// UnmarshalJSON 兼容 -session.json 的 snake_case 字段
func (s *SessionInfo) UnmarshalJSON(data []byte) error {
	type alias struct {
		ID         string `json:"id"`
		Title      string `json:"title"`
		WorkingDir string `json:"working_dir"`
		CreatedAt  int64  `json:"created_at"`
		UpdatedAt  int64  `json:"updated_at"`
	}
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	s.ID = a.ID
	s.Title = a.Title
	s.WorkingDir = a.WorkingDir
	s.CreatedAt = a.CreatedAt
	s.UpdatedAt = a.UpdatedAt
	return nil
}

// SessionData 会话完整数据（用于备份）
type SessionData struct {
	SessionInfo
	Messages []ChatMessage `json:"messages"`
}

// ChatMessage 单条会话消息
type ChatMessage struct {
	ID         string     `json:"id"`
	Role       string     `json:"role"`
	SessionID  string     `json:"sessionId"`
	Text       string     `json:"text"`
	Reasoning  string     `json:"reasoning,omitempty"`
	Images     []string   `json:"images,omitempty"`
	ToolCalls  []ToolCall `json:"toolCalls,omitempty"`
	ToolResult string     `json:"toolResult,omitempty"`
	CreatedAt  int64      `json:"createdAt"`
	Raw        string     `json:"raw,omitempty"`
}

// ToolCall 工具调用
type ToolCall struct {
	Name string `json:"name"`
	Args string `json:"args"`
}

// transcriptDir 返回 qodercli 的会话转录根目录
func (m *QoderSessionManager) transcriptDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".qoder", "projects")
}

// legacyProjectsDir 返回旧的会话项目目录（用于读取标题元数据）
func (m *QoderSessionManager) legacyProjectsDir() string {
	return filepath.Join(m.basePath, "SharedClientCache", "cli", "projects")
}

// workdirFromDirName 将目录名 -Users-ray-Downloads-SameMoon 还原为路径
func workdirFromDirName(name string) string {
	if !strings.HasPrefix(name, "-") {
		return name
	}
	return "/" + strings.ReplaceAll(strings.TrimPrefix(name, "-"), "-", "/")
}

// ListSessions 列出所有会话（按更新时间倒序）
func (m *QoderSessionManager) ListSessions() ([]SessionInfo, error) {
	dirs, err := os.ReadDir(m.transcriptDir())
	if err != nil {
		return nil, err
	}

	var sessions []SessionInfo
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		transDir := filepath.Join(m.transcriptDir(), dir.Name(), "transcript")
		files, err := os.ReadDir(transDir)
		if err != nil {
			continue
		}

		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".jsonl") {
				continue
			}

			sessionID := strings.TrimSuffix(f.Name(), ".jsonl")
			info := m.scanTranscript(filepath.Join(transDir, f.Name()), sessionID)
			if info.MessageCount == 0 {
				continue
			}
			info.WorkingDir = workdirFromDirName(dir.Name())

			// 标题：优先从旧 -session.json 读取
			if legacy, err := os.ReadFile(filepath.Join(m.legacyProjectsDir(), sessionID+"-session.json")); err == nil {
				var li SessionInfo
				if err := json.Unmarshal(legacy, &li); err == nil && li.Title != "" {
					info.Title = li.Title
				}
			}
			// 兜底：首条用户消息前 50 字
			if info.Title == "" {
				info.Title = m.firstUserTitle(filepath.Join(transDir, f.Name()))
			}
			if info.Title == "" {
				info.Title = "未命名会话"
			}

			sessions = append(sessions, info)
		}
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt > sessions[j].UpdatedAt
	})
	return sessions, nil
}

// scanTranscript 扫描 transcript 统计时间与消息数
func (m *QoderSessionManager) scanTranscript(path, sessionID string) SessionInfo {
	info := SessionInfo{ID: sessionID}
	f, err := os.Open(path)
	if err != nil {
		return info
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 64*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var ev struct {
			Type      string `json:"type"`
			Timestamp string `json:"timestamp"`
		}
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}
		if ev.Type != "user" && ev.Type != "assistant" {
			continue
		}
		info.MessageCount++
		ts := parseRFC3339Millis(ev.Timestamp)
		if ts > 0 {
			if info.CreatedAt == 0 || ts < info.CreatedAt {
				info.CreatedAt = ts
			}
			if ts > info.UpdatedAt {
				info.UpdatedAt = ts
			}
		}
	}
	return info
}

// firstUserTitle 提取首条用户消息作为标题
func (m *QoderSessionManager) firstUserTitle(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 64*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var ev struct {
			Type    string `json:"type"`
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}
		if err := json.Unmarshal([]byte(line), &ev); err != nil || ev.Type != "user" {
			continue
		}
		text := strings.TrimSpace(ev.Message.Content)
		if text == "" {
			continue
		}
		runes := []rune(text)
		if len(runes) > 50 {
			return string(runes[:50]) + "..."
		}
		return text
	}
	return ""
}

func parseRFC3339Millis(s string) int64 {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return 0
	}
	return t.UnixMilli()
}

// GetSessionMessages 读取会话完整消息（transcript 格式）
func (m *QoderSessionManager) GetSessionMessages(sessionID string) ([]ChatMessage, error) {
	path, err := m.findTranscript(sessionID)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var messages []ChatMessage
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 64*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var raw struct {
			Type      string `json:"type"`
			SessionID string `json:"sessionId"`
			UUID      string `json:"uuid"`
			Timestamp string `json:"timestamp"`
			Message   struct {
				Role    string          `json:"role"`
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}

		msg := ChatMessage{
			ID:        raw.UUID,
			SessionID: raw.SessionID,
			CreatedAt: parseRFC3339Millis(raw.Timestamp),
			Raw:       line,
		}

		switch raw.Type {
		case "user":
			msg.Role = "user"
			var content string
			if err := json.Unmarshal(raw.Message.Content, &content); err == nil {
				msg.Text = content
				messages = append(messages, msg)
				continue
			}
			// 工具结果：content 为数组
			var parts []struct {
				Type      string `json:"type"`
				Content   string `json:"content"`
				ToolUseID string `json:"tool_use_id"`
			}
			if err := json.Unmarshal(raw.Message.Content, &parts); err == nil {
				for _, p := range parts {
					if p.ToolUseID != "" {
						msg.Role = "tool"
						msg.ToolResult = truncateStr(p.Content, 2000)
						messages = append(messages, msg)
					}
				}
			}
		case "assistant":
			msg.Role = "assistant"
			var parts []map[string]json.RawMessage
			if err := json.Unmarshal(raw.Message.Content, &parts); err != nil {
				continue
			}
			var textParts, thinkingParts []string
			for _, p := range parts {
				var ptype string
				json.Unmarshal(p["type"], &ptype)
				switch ptype {
				case "thinking":
					if s, ok := jsonRawString(p["thinking"]); ok {
						thinkingParts = append(thinkingParts, s)
					}
				case "text":
					if s, ok := jsonRawString(p["text"]); ok {
						textParts = append(textParts, s)
					}
				case "toolCall", "tool_use":
					var tc struct {
						Name  string `json:"name"`
						Input string `json:"input"`
					}
					json.Unmarshal(p["name"], &tc.Name)
					var input json.RawMessage
					if err := json.Unmarshal(p["input"], &input); err == nil {
						if s, ok := jsonRawString(input); ok {
							tc.Input = s
						} else {
							tc.Input = string(input)
						}
					}
					if tc.Name != "" {
						msg.ToolCalls = append(msg.ToolCalls, ToolCall{Name: tc.Name, Args: tc.Input})
					}
				}
			}
			msg.Text = strings.Join(textParts, "\n")
			msg.Reasoning = strings.Join(thinkingParts, "\n")
			messages = append(messages, msg)
		}
	}
	return messages, nil
}

func jsonRawString(raw json.RawMessage) (string, bool) {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, true
	}
	return "", false
}

// findTranscript 查找 sessionID 对应的 transcript 文件
func (m *QoderSessionManager) findTranscript(sessionID string) (string, error) {
	dirs, err := os.ReadDir(m.transcriptDir())
	if err != nil {
		return "", err
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		p := filepath.Join(m.transcriptDir(), dir.Name(), "transcript", sessionID+".jsonl")
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("会话 %s 不存在", sessionID)
}

// listTranscriptPaths 列出全部 transcript 文件路径（备份用）
func (m *QoderSessionManager) listTranscriptPaths() []string {
	var paths []string
	dirs, err := os.ReadDir(m.transcriptDir())
	if err != nil {
		return paths
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		files, err := os.ReadDir(filepath.Join(m.transcriptDir(), dir.Name(), "transcript"))
		if err != nil {
			continue
		}
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".jsonl") {
				paths = append(paths, filepath.Join(m.transcriptDir(), dir.Name(), "transcript", f.Name()))
			}
		}
	}
	return paths
}

func truncateStr(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}
