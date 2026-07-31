package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"qoder-gui/qoder"
)

// App 应用主结构
type App struct {
	ctx     context.Context
	manager *qoder.QoderSessionManager
}

// WorkspaceInfo 工作区信息
type WorkspaceInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	ChatCount int    `json:"chatCount"`
}

// ChatInfo 会话信息
type ChatInfo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Timestamp int64  `json:"timestamp"`
	SessionID string `json:"sessionId"`
	TimeStr   string `json:"timeStr"`
}

// BackupInfo 备份信息
type BackupInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	ModTime int64  `json:"modTime"`
	TimeStr string `json:"timeStr"`
}

// BackupResult 备份结果
type BackupResult struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	BackupPath   string `json:"backupPath,omitempty"`
	Count        int    `json:"count"`
	SuccessCount int    `json:"successCount"`
}

// RestoreResult 恢复结果
type RestoreResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// NewApp 创建应用
func NewApp() *App {
	return &App{}
}

// startup 启动时调用
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.manager = qoder.NewQoderSessionManager()
}

// GetUserID 获取当前用户ID
func (a *App) GetUserID() string {
	userID, _ := a.manager.GetUserID()
	return userID
}

// GetWorkspaces 获取工作区列表
func (a *App) GetWorkspaces() []WorkspaceInfo {
	workspaces, err := a.manager.ListWorkspaces()
	if err != nil {
		return []WorkspaceInfo{}
	}

	var result []WorkspaceInfo
	for id, path := range workspaces {
		name := filepath.Base(path)
		if name == "" || name == "/" {
			name = "未命名工作区"
		}

		history, _ := a.manager.GetWorkspaceChatHistory(id)

		result = append(result, WorkspaceInfo{
			ID:        id,
			Name:      name,
			Path:      path,
			ChatCount: len(history),
		})
	}

	// 按名称排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// GetWorkspaceChats 获取工作区的会话列表
func (a *App) GetWorkspaceChats(workspaceID string) []ChatInfo {
	history, err := a.manager.GetWorkspaceChatHistory(workspaceID)
	if err != nil {
		return []ChatInfo{}
	}

	var result []ChatInfo
	for _, h := range history {
		timestamp := time.UnixMilli(h.Timestamp)
		result = append(result, ChatInfo{
			ID:        h.ID,
			Title:     h.Title,
			Timestamp: h.Timestamp,
			SessionID: h.SessionID,
			TimeStr:   timestamp.Format("2006-01-02 15:04:05"),
		})
	}

	// 按时间倒序排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp > result[j].Timestamp
	})

	return result
}

// BackupAll 备份所有工作区
func (a *App) BackupAll() BackupResult {
	workspaces, err := a.manager.ListWorkspaces()
	if err != nil {
		return BackupResult{
			Success: false,
			Message: "获取工作区列表失败",
		}
	}

	var allBackups []qoder.SessionBackup
	successCount := 0

	for workspaceID, workspacePath := range workspaces {
		backup, err := a.manager.BackupWorkspace(workspaceID, workspacePath)
		if err != nil {
			continue
		}
		allBackups = append(allBackups, *backup)
		successCount++
	}

	if len(allBackups) == 0 {
		return BackupResult{
			Success: false,
			Message: "没有成功备份任何工作区",
		}
	}

	// 保存备份
	backupDir := filepath.Join(os.Getenv("HOME"), "Documents", "qoder-backups")
	os.MkdirAll(backupDir, 0755)
	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("qoder-backup-all-%s.json", timestamp))

	data, _ := json.MarshalIndent(allBackups, "", "  ")
	os.WriteFile(backupPath, data, 0644)

	return BackupResult{
		Success:      true,
		Message:      fmt.Sprintf("成功备份 %d 个工作区", successCount),
		BackupPath:   backupPath,
		Count:        len(allBackups),
		SuccessCount: successCount,
	}
}

// GetBackups 获取备份列表
func (a *App) GetBackups() []BackupInfo {
	backupDir := filepath.Join(os.Getenv("HOME"), "Documents", "qoder-backups")
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return []BackupInfo{}
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !filepath.HasPrefix(entry.Name(), "qoder-backup") {
			continue
		}

		info, _ := entry.Info()
		modTime := info.ModTime()

		backups = append(backups, BackupInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(backupDir, entry.Name()),
			Size:    info.Size(),
			ModTime: modTime.Unix(),
			TimeStr: modTime.Format("2006-01-02 15:04:05"),
		})
	}

	// 按修改时间倒序排序
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].ModTime > backups[j].ModTime
	})

	return backups
}

// RestoreBackup 恢复备份
func (a *App) RestoreBackup(backupPath string) RestoreResult {
	if err := a.manager.RestoreBackup(backupPath); err != nil {
		return RestoreResult{
			Success: false,
			Message: "恢复失败",
			Error:   err.Error(),
		}
	}

	return RestoreResult{
		Success: true,
		Message: "恢复成功！请重启Qoder查看效果",
	}
}

// ExportSessions 导出会话为文本
func (a *App) ExportSessions() string {
	workspaces, err := a.manager.ListWorkspaces()
	if err != nil {
		return "导出失败: " + err.Error()
	}

	backupDir := filepath.Join(os.Getenv("HOME"), "Documents", "qoder-backups")
	timestamp := time.Now().Format("20060102-150405")
	exportPath := filepath.Join(backupDir, fmt.Sprintf("qoder-export-%s.md", timestamp))

	var content string
	content += "# Qoder 会话导出\n\n"
	content += fmt.Sprintf("导出时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	for workspaceID, workspacePath := range workspaces {
		backup, err := a.manager.BackupWorkspace(workspaceID, workspacePath)
		if err != nil {
			continue
		}

		content += fmt.Sprintf("## 工作区: %s\n", backup.WorkspacePath)
		content += fmt.Sprintf("会话数: %d\n\n", len(backup.ChatHistory))

		for i, h := range backup.ChatHistory {
			timestamp := time.UnixMilli(h.Timestamp).Format("2006-01-02 15:04:05")
			content += fmt.Sprintf("### %d. %s\n", i+1, h.Title)
			content += fmt.Sprintf("- 时间: %s\n", timestamp)
			content += fmt.Sprintf("- 会话ID: %s\n\n", h.SessionID)
		}
		content += "\n---\n\n"
	}

	os.WriteFile(exportPath, []byte(content), 0644)

	return exportPath
}

// DeleteBackup 删除备份
func (a *App) DeleteBackup(backupPath string) bool {
	err := os.Remove(backupPath)
	return err == nil
}

// OpenBackupFolder 打开备份文件夹
func (a *App) OpenBackupFolder() string {
	backupDir := filepath.Join(os.Getenv("HOME"), "Documents", "qoder-backups")
	return backupDir
}
