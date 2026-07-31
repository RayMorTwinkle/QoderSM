package qoder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// WebServer Web服务器
type WebServer struct {
	manager *QoderSessionManager
	port    int
}

// NewWebServer 创建Web服务器
func NewWebServer(port int) *WebServer {
	return &WebServer{
		manager: NewQoderSessionManager(),
		port:    port,
	}
}

// Start 启动Web服务器
func (s *WebServer) Start() error {
	// 查找web目录（支持开发环境和.app包）
	webDir := s.findWebDir()
	if webDir == "" {
		webDir = "."
	}

	// 静态文件服务
	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)

	// API路由
	http.HandleFunc("/api/workspaces", s.handleWorkspaces)
	http.HandleFunc("/api/workspace/", s.handleWorkspace)
	http.HandleFunc("/api/backup", s.handleBackup)
	http.HandleFunc("/api/restore", s.handleRestore)
	http.HandleFunc("/api/export", s.handleExport)
	http.HandleFunc("/api/backups", s.handleBackups)

	addr := fmt.Sprintf("localhost:%d", s.port)
	fmt.Printf("服务器已启动: http://%s\n", addr)

	// 自动打开浏览器
	time.Sleep(500 * time.Millisecond)
	openBrowser(addr)

	return http.ListenAndServe(addr, nil)
}

func (s *WebServer) handleWorkspaces(w http.ResponseWriter, r *http.Request) {
	workspaces, err := s.manager.ListWorkspaces()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	type WorkspaceInfo struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Path       string `json:"path"`
		ChatCount  int    `json:"chatCount"`
	}

	var result []WorkspaceInfo
	for id, path := range workspaces {
		name := filepath.Base(path)
		if name == "" || name == "/" {
			name = "未命名工作区"
		}

		history, _ := s.manager.GetWorkspaceChatHistory(id)

		result = append(result, WorkspaceInfo{
			ID:        id,
			Name:      name,
			Path:      path,
			ChatCount: len(history),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *WebServer) handleWorkspace(w http.ResponseWriter, r *http.Request) {
	// 从路径提取workspace ID
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", 400)
		return
	}
	workspaceID := parts[3]

	// 获取会话历史
	history, err := s.manager.GetWorkspaceChatHistory(workspaceID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// 获取聊天视图
	views, _ := s.manager.GetWorkspaceChatViews(workspaceID)

	// 获取聊天标签
	tabs, _ := s.manager.GetWorkspaceChatTabs(workspaceID)

	result := map[string]interface{}{
		"history": history,
		"views":   views,
		"tabs":    tabs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *WebServer) handleBackup(w http.ResponseWriter, r *http.Request) {
	workspaces, err := s.manager.ListWorkspaces()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var allBackups []SessionBackup
	successCount := 0

	for workspaceID, workspacePath := range workspaces {
		backup, err := s.manager.BackupWorkspace(workspaceID, workspacePath)
		if err != nil {
			continue
		}
		allBackups = append(allBackups, *backup)
		successCount++
	}

	// 保存备份
	backupDir := getDefaultBackupDir()
	os.MkdirAll(backupDir, 0755)
	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("qoder-backup-all-%s.json", timestamp))

	data, _ := json.MarshalIndent(allBackups, "", "  ")
	os.WriteFile(backupPath, data, 0644)

	result := map[string]interface{}{
		"success":      true,
		"backupPath":   backupPath,
		"count":        len(allBackups),
		"successCount": successCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *WebServer) handleRestore(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req struct {
		BackupPath string `json:"backupPath"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if err := s.manager.RestoreBackup(req.BackupPath); err != nil {
		result := map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	result := map[string]interface{}{
		"success": true,
		"message": "恢复成功，请重启Qoder",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *WebServer) handleExport(w http.ResponseWriter, r *http.Request) {
	workspaces, err := s.manager.ListWorkspaces()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var sb strings.Builder
	sb.WriteString("# Qoder 会话导出\n\n")
	sb.WriteString(fmt.Sprintf("导出时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	for workspaceID, workspacePath := range workspaces {
		backup, err := s.manager.BackupWorkspace(workspaceID, workspacePath)
		if err != nil {
			continue
		}

		sb.WriteString(fmt.Sprintf("## 工作区: %s\n", backup.WorkspacePath))
		sb.WriteString(fmt.Sprintf("会话数: %d\n\n", len(backup.ChatHistory)))

		for i, h := range backup.ChatHistory {
			timestamp := time.UnixMilli(h.Timestamp).Format("2006-01-02 15:04:05")
			sb.WriteString(fmt.Sprintf("### %d. %s\n", i+1, h.Title))
			sb.WriteString(fmt.Sprintf("- 时间: %s\n", timestamp))
			sb.WriteString(fmt.Sprintf("- 会话ID: %s\n\n", h.SessionID))
		}
		sb.WriteString("\n---\n\n")
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=qoder-export-%s.md", time.Now().Format("20060102-150405")))
	w.Write([]byte(sb.String()))
}

func (s *WebServer) handleBackups(w http.ResponseWriter, r *http.Request) {
	backupDir := getDefaultBackupDir()
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	type BackupInfo struct {
		Name     string    `json:"name"`
		Path     string    `json:"path"`
		Size     int64     `json:"size"`
		ModTime  time.Time `json:"modTime"`
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasPrefix(entry.Name(), "qoder-backup") {
			continue
		}

		info, _ := entry.Info()
		backups = append(backups, BackupInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(backupDir, entry.Name()),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}

	// 按修改时间排序
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].ModTime.After(backups[j].ModTime)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(backups)
}

func openBrowser(url string) {
	cmd := "open"
	exec.Command(cmd, url).Start()
}

func getDefaultBackupDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "Documents", "qoder-backups")
}

// findWebDir 查找web目录
func (s *WebServer) findWebDir() string {
	// 首先尝试当前目录的web文件夹（开发环境）
	if _, err := os.Stat("web/index.html"); err == nil {
		return "web"
	}

	// 尝试相对于可执行文件的路径（.app包）
	execPath, err := os.Executable()
	if err == nil {
		// 在.app包中，路径是 .../QoderSessionManager.app/Contents/MacOS/qoder-web
		// 我们需要 .../QoderSessionManager.app/Contents/Resources/web
		if filepath.Base(filepath.Dir(filepath.Dir(execPath))) == "Contents" {
			resourcesDir := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(execPath))), "Resources", "web")
			if _, err := os.Stat(filepath.Join(resourcesDir, "index.html")); err == nil {
				return resourcesDir
			}
		}
	}

	// 尝试源码目录
	sourceDir := "/Users/lucky/Documents/WorkSapce/QoderSM/web"
	if _, err := os.Stat(filepath.Join(sourceDir, "index.html")); err == nil {
		return sourceDir
	}

	return ""
}
