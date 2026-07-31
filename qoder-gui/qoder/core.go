package qoder

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// QoderSessionManager 管理Qoder的会话
type QoderSessionManager struct {
	basePath string
}

// ChatHistory 聊天历史记录
type ChatHistory struct {
	ID        string      `json:"id"`
	Title     string      `json:"title"`
	Context   []Context   `json:"context"`
	Timestamp int64       `json:"timestamp"`
	SessionID string      `json:"sessionId"`
}

// Context 对话上下文
type Context struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ChatViews 聊天视图配置 (可以是对象或直接数组)
type ChatViews struct {
	Views []ChatView `json:"views,omitempty"`
}

// UnmarshalJSON 自定义JSON解析，支持直接数组和对象两种格式
func (cv *ChatViews) UnmarshalJSON(data []byte) error {
	// 尝试解析为直接数组
	var directArray []ChatView
	if err := json.Unmarshal(data, &directArray); err == nil {
		cv.Views = directArray
		return nil
	}

	// 尝试解析为对象
	type chatViewsAlias ChatViews
	alias := &struct {
		Views []ChatView `json:"views"`
	}{}
	if err := json.Unmarshal(data, &alias); err == nil {
		cv.Views = alias.Views
		return nil
	}

	return fmt.Errorf("无法解析ChatViews数据")
}

// MarshalJSON 自定义JSON序列化
func (cv ChatViews) MarshalJSON() ([]byte, error) {
	return json.Marshal(cv.Views)
}

// ChatView 单个聊天视图
type ChatView struct {
	ID                 string `json:"id"`
	ViewID             string `json:"viewId"`
	SessionID          string `json:"sessionId"`
	Title              string `json:"title"`
	Location           string `json:"location"`
	Active             bool   `json:"active"`
	SessionType        string `json:"sessionType"`
	TargetRemoteAuthority interface{} `json:"targetRemoteAuthority"`
}

// ChatTabs 聊天标签页配置
type ChatTabs struct {
	Tabs        []ChatTab `json:"tabs"`
	ActiveTabID string    `json:"activeTabId"`
}

// ChatTab 单个聊天标签
type ChatTab struct {
	TabID    string `json:"tabId"`
	SessionID string `json:"sessionId"`
	Title    string `json:"title"`
}

// SessionBackup 会话备份数据
type SessionBackup struct {
	BackupTime   time.Time            `json:"backupTime"`
	WorkspaceID  string              `json:"workspaceId"`
	WorkspacePath string             `json:"workspacePath"`
	UserID       string              `json:"userId"`
	ChatHistory  []ChatHistory       `json:"chatHistory"`
	ChatViews    ChatViews           `json:"chatViews"`
	ChatTabs     ChatTabs            `json:"chatTabs"`
}

// NewQoderSessionManager 创建会话管理器
func NewQoderSessionManager() *QoderSessionManager {
	homeDir, _ := os.UserHomeDir()
	return &QoderSessionManager{
		basePath: filepath.Join(homeDir, "Library", "Application Support", "Qoder"),
	}
}

// GetUserID 获取当前用户ID
func (m *QoderSessionManager) GetUserID() (string, error) {
	data, err := os.ReadFile(filepath.Join(m.basePath, ".auth", "id"))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ListWorkspaces 列出所有工作区
func (m *QoderSessionManager) ListWorkspaces() (map[string]string, error) {
	workspacePath := filepath.Join(m.basePath, "User", "workspaceStorage")
	entries, err := os.ReadDir(workspacePath)
	if err != nil {
		return nil, err
	}

	workspaces := make(map[string]string)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		workspaceID := entry.Name()
		workspaceJSON := filepath.Join(workspacePath, workspaceID, "workspace.json")
		data, err := os.ReadFile(workspaceJSON)
		if err != nil {
			continue
		}

		var ws struct {
			Folder string `json:"folder"`
		}
		if err := json.Unmarshal(data, &ws); err == nil {
			workspaces[workspaceID] = ws.Folder
		}
	}
	return workspaces, nil
}

// GetWorkspaceChatHistory 获取工作区的聊天历史
func (m *QoderSessionManager) GetWorkspaceChatHistory(workspaceID string) ([]ChatHistory, error) {
	dbPath := filepath.Join(m.basePath, "User", "globalStorage", "state.vscdb")
	return m.getChatHistoryFromDB(dbPath, workspaceID)
}

// getChatHistoryFromDB 从数据库获取聊天历史
func (m *QoderSessionManager) getChatHistoryFromDB(dbPath, workspaceID string) ([]ChatHistory, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	key := fmt.Sprintf("lingma.chat.localHistory.%s", workspaceID)
	var value string
	err = db.QueryRow("SELECT value FROM ItemTable WHERE key = ?", key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return []ChatHistory{}, nil
		}
		return nil, err
	}

	var history []ChatHistory
	if err := json.Unmarshal([]byte(value), &history); err != nil {
		return nil, err
	}
	return history, nil
}

// GetWorkspaceChatViews 获取工作区的聊天视图
func (m *QoderSessionManager) GetWorkspaceChatViews(workspaceID string) (ChatViews, error) {
	dbPath := filepath.Join(m.basePath, "User", "workspaceStorage", workspaceID, "state.vscdb")
	return m.getChatViewsFromDB(dbPath)
}

// getChatViewsFromDB 从数据库获取聊天视图
func (m *QoderSessionManager) getChatViewsFromDB(dbPath string) (ChatViews, error) {
	var views ChatViews

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return views, err
	}
	defer db.Close()

	var value string
	err = db.QueryRow("SELECT value FROM ItemTable WHERE key = ?", "aicoding.chat.views").Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return views, nil
		}
		return views, err
	}

	if err := json.Unmarshal([]byte(value), &views); err != nil {
		return views, err
	}
	return views, nil
}

// GetWorkspaceChatTabs 获取工作区的聊天标签
func (m *QoderSessionManager) GetWorkspaceChatTabs(workspaceID string) (ChatTabs, error) {
	dbPath := filepath.Join(m.basePath, "User", "workspaceStorage", workspaceID, "state.vscdb")
	return m.getChatTabsFromDB(dbPath)
}

// getChatTabsFromDB 从数据库获取聊天标签
func (m *QoderSessionManager) getChatTabsFromDB(dbPath string) (ChatTabs, error) {
	var tabs ChatTabs

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return tabs, err
	}
	defer db.Close()

	var value string
	err = db.QueryRow("SELECT value FROM ItemTable WHERE key = ?", "aicoding.chat.tabs").Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return tabs, nil
		}
		return tabs, err
	}

	if err := json.Unmarshal([]byte(value), &tabs); err != nil {
		return tabs, err
	}
	return tabs, nil
}

// BackupWorkspace 备份工作区会话
func (m *QoderSessionManager) BackupWorkspace(workspaceID, workspacePath string) (*SessionBackup, error) {
	userID, _ := m.GetUserID()

	history, err := m.GetWorkspaceChatHistory(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("获取聊天历史失败: %w", err)
	}

	views, err := m.GetWorkspaceChatViews(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("获取聊天视图失败: %w", err)
	}

	tabs, err := m.GetWorkspaceChatTabs(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("获取聊天标签失败: %w", err)
	}

	backup := &SessionBackup{
		BackupTime:    time.Now(),
		WorkspaceID:   workspaceID,
		WorkspacePath: workspacePath,
		UserID:        userID,
		ChatHistory:   history,
		ChatViews:     views,
		ChatTabs:      tabs,
	}

	return backup, nil
}

// SaveBackup 保存备份到文件
func (m *QoderSessionManager) SaveBackup(backup *SessionBackup, backupPath string) error {
	os.MkdirAll(filepath.Dir(backupPath), 0755)

	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(backupPath, data, 0644)
}

// RestoreBackup 恢复备份
func (m *QoderSessionManager) RestoreBackup(backupPath string) error {
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return err
	}

	// 先尝试解析为数组格式（BackupAll 生成的格式）
	var backups []SessionBackup
	if err := json.Unmarshal(data, &backups); err == nil {
		for _, backup := range backups {
			if err := m.restoreSingleBackup(backup); err != nil {
				log.Printf("警告: 恢复工作区 %s 失败: %v", backup.WorkspaceID, err)
			}
		}
		return nil
	}

	// 再尝试解析为单个对象格式
	var backup SessionBackup
	if err := json.Unmarshal(data, &backup); err != nil {
		return err
	}
	return m.restoreSingleBackup(backup)
}

// restoreSingleBackup 恢复单个工作区备份
func (m *QoderSessionManager) restoreSingleBackup(backup SessionBackup) error {
	// 恢复聊天历史到全局存储
	if err := m.restoreChatHistory(backup.WorkspaceID, backup.ChatHistory); err != nil {
		log.Printf("警告: 恢复聊天历史失败: %v", err)
	}

	// 恢复聊天视图
	if err := m.restoreChatViews(backup.WorkspaceID, backup.ChatViews); err != nil {
		log.Printf("警告: 恢复聊天视图失败: %v", err)
	}

	// 恢复聊天标签
	if err := m.restoreChatTabs(backup.WorkspaceID, backup.ChatTabs); err != nil {
		log.Printf("警告: 恢复聊天标签失败: %v", err)
	}

	return nil
}

// restoreChatHistory 恢复聊天历史
func (m *QoderSessionManager) restoreChatHistory(workspaceID string, history []ChatHistory) error {
	dbPath := filepath.Join(m.basePath, "User", "globalStorage", "state.vscdb")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	key := fmt.Sprintf("lingma.chat.localHistory.%s", workspaceID)
	value, _ := json.Marshal(history)

	_, err = db.Exec("INSERT OR REPLACE INTO ItemTable (key, value) VALUES (?, ?)", key, string(value))
	return err
}

// restoreChatViews 恢复聊天视图
func (m *QoderSessionManager) restoreChatViews(workspaceID string, views ChatViews) error {
	dbPath := filepath.Join(m.basePath, "User", "workspaceStorage", workspaceID, "state.vscdb")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	value, _ := json.Marshal(views)

	_, err = db.Exec("INSERT OR REPLACE INTO ItemTable (key, value) VALUES (?, ?)", "aicoding.chat.views", string(value))
	return err
}

// restoreChatTabs 恢复聊天标签
func (m *QoderSessionManager) restoreChatTabs(workspaceID string, tabs ChatTabs) error {
	dbPath := filepath.Join(m.basePath, "User", "workspaceStorage", workspaceID, "state.vscdb")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	value, _ := json.Marshal(tabs)

	_, err = db.Exec("INSERT OR REPLACE INTO ItemTable (key, value) VALUES (?, ?)", "aicoding.chat.tabs", string(value))
	return err
}

// ListBackups 列出所有备份
func (m *QoderSessionManager) ListBackups(backupDir string) ([]SessionBackup, error) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, err
	}

	var backups []SessionBackup
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(backupDir, entry.Name()))
		if err != nil {
			continue
		}

		var backup SessionBackup
		if err := json.Unmarshal(data, &backup); err != nil {
			continue
		}

		backups = append(backups, backup)
	}

	return backups, nil
}
