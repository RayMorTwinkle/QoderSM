package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"qoder-sm/pkg/qoder"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	manager := qoder.NewQoderSessionManager()
	command := os.Args[1]

	switch command {
	case "list":
		listWorkspaces(manager)
	case "backup":
		if len(os.Args) < 3 {
			fmt.Println("请指定工作区ID或使用 'all' 备份所有工作区")
			return
		}
		backupWorkspace(manager, os.Args[2])
	case "restore":
		if len(os.Args) < 3 {
			fmt.Println("请指定备份文件路径")
			return
		}
		restoreWorkspace(manager, os.Args[2])
	case "show":
		if len(os.Args) < 3 {
			fmt.Println("请指定工作区ID")
			return
		}
		showWorkspace(manager, os.Args[2])
	case "export":
		if len(os.Args) < 3 {
			fmt.Println("请指定工作区ID或使用 'all' 导出所有工作区")
			return
		}
		exportWorkspace(manager, os.Args[2])
	case "list-backups":
		if len(os.Args) < 3 {
			listBackups(manager, getDefaultBackupDir())
		} else {
			listBackups(manager, os.Args[2])
		}
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Qoder会话管理工具")
	fmt.Println("\n用法:")
	fmt.Println("  qoder-sm list                          - 列出所有工作区")
	fmt.Println("  qoder-sm backup <workspace-id|all>    - 备份指定工作区或所有工作区的会话")
	fmt.Println("  qoder-sm restore <backup-file>         - 从备份文件恢复会话")
	fmt.Println("  qoder-sm show <workspace-id>           - 显示工作区的会话详情")
	fmt.Println("  qoder-sm export <workspace-id|all>     - 导出会话为可读格式")
	fmt.Println("  qoder-sm list-backups [backup-dir]     - 列出所有备份")
	fmt.Println("\n示例:")
	fmt.Println("  qoder-sm backup all")
	fmt.Println("  qoder-sm restore ~/Documents/qoder-backup-20260524.json")
	fmt.Println("  qoder-sm show 5a1eb88dae06d83ff14cad86bae758a4")
}

func listWorkspaces(manager *qoder.QoderSessionManager) {
	workspaces, err := manager.ListWorkspaces()
	if err != nil {
		fmt.Printf("获取工作区列表失败: %v\n", err)
		return
	}

	userID, _ := manager.GetUserID()

	fmt.Printf("当前用户ID: %s\n\n", userID)
	fmt.Println("工作区列表:")
	fmt.Println("----------------------------------------")

	for id, path := range workspaces {
		fmt.Printf("ID: %s\n", id)
		fmt.Printf("路径: %s\n", path)

		// 获取会话数量
		history, err := manager.GetWorkspaceChatHistory(id)
		if err == nil {
			fmt.Printf("会话数: %d\n", len(history))
		}

		fmt.Println("----------------------------------------")
	}
}

func backupWorkspace(manager *qoder.QoderSessionManager, workspaceID string) {
	if workspaceID == "all" {
		backupAll(manager)
		return
	}

	workspaces, err := manager.ListWorkspaces()
	if err != nil {
		fmt.Printf("获取工作区列表失败: %v\n", err)
		return
	}

	workspacePath, exists := workspaces[workspaceID]
	if !exists {
		fmt.Printf("工作区 %s 不存在\n", workspaceID)
		return
	}

	fmt.Printf("正在备份工作区: %s\n", workspacePath)

	backup, err := manager.BackupWorkspace(workspaceID, workspacePath)
	if err != nil {
		fmt.Printf("备份失败: %v\n", err)
		return
	}

	backupDir := getDefaultBackupDir()
	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("qoder-backup-%s.json", timestamp))

	if err := manager.SaveBackup(backup, backupPath); err != nil {
		fmt.Printf("保存备份失败: %v\n", err)
		return
	}

	fmt.Printf("备份已保存到: %s\n", backupPath)
	fmt.Printf("  - 会话历史: %d 条\n", len(backup.ChatHistory))
	fmt.Printf("  - 聊天视图: %d 个\n", len(backup.ChatViews.Views))
	fmt.Printf("  - 聊天标签: %d 个\n", len(backup.ChatTabs.Tabs))
}

func backupAll(manager *qoder.QoderSessionManager) {
	workspaces, err := manager.ListWorkspaces()
	if err != nil {
		fmt.Printf("获取工作区列表失败: %v\n", err)
		return
	}

	backupDir := getDefaultBackupDir()
	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("qoder-backup-all-%s.json", timestamp))

	var allBackups []qoder.SessionBackup

	fmt.Printf("正在备份 %d 个工作区...\n", len(workspaces))

	for workspaceID, workspacePath := range workspaces {
		fmt.Printf("  - %s\n", workspacePath)

		backup, err := manager.BackupWorkspace(workspaceID, workspacePath)
		if err != nil {
			fmt.Printf("    失败: %v\n", err)
			continue
		}

		allBackups = append(allBackups, *backup)
	}

	// 保存合并的备份
	data, err := json.MarshalIndent(allBackups, "", "  ")
	if err != nil {
		fmt.Printf("序列化备份失败: %v\n", err)
		return
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		fmt.Printf("保存备份失败: %v\n", err)
		return
	}

	fmt.Printf("\n备份已保存到: %s\n", backupPath)
	fmt.Printf("成功备份 %d 个工作区\n", len(allBackups))
}

func restoreWorkspace(manager *qoder.QoderSessionManager, backupPath string) {
	fmt.Printf("正在从备份恢复: %s\n", backupPath)

	// 检查是单个备份还是批量备份
	data, err := os.ReadFile(backupPath)
	if err != nil {
		fmt.Printf("读取备份文件失败: %v\n", err)
		return
	}

	// 尝试解析为单个备份
	var singleBackup qoder.SessionBackup
	if err := json.Unmarshal(data, &singleBackup); err == nil && singleBackup.BackupTime.Unix() > 0 {
		if err := manager.RestoreBackup(backupPath); err != nil {
			fmt.Printf("恢复失败: %v\n", err)
			return
		}
		fmt.Printf("成功恢复工作区: %s\n", singleBackup.WorkspacePath)
		fmt.Printf("  - 恢复 %d 条会话历史\n", len(singleBackup.ChatHistory))
		return
	}

	// 尝试解析为批量备份
	var allBackups []qoder.SessionBackup
	if err := json.Unmarshal(data, &allBackups); err != nil {
		fmt.Printf("解析备份文件失败: %v\n", err)
		return
	}

	fmt.Printf("批量备份包含 %d 个工作区\n", len(allBackups))
	for i := range allBackups {
		// 临时保存单个备份
		tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("qoder-temp-backup-%d.json", i))
		tempData, _ := json.MarshalIndent(allBackups[i], "", "  ")
		os.WriteFile(tempFile, tempData, 0644)

		if err := manager.RestoreBackup(tempFile); err != nil {
			fmt.Printf("恢复工作区 %s 失败: %v\n", allBackups[i].WorkspacePath, err)
			continue
		}

		fmt.Printf("  - 成功恢复: %s (%d 条会话)\n",
			allBackups[i].WorkspacePath,
			len(allBackups[i].ChatHistory))

		os.Remove(tempFile)
	}

	fmt.Println("\n恢复完成！请重启Qoder以查看恢复的会话。")
}

func showWorkspace(manager *qoder.QoderSessionManager, workspaceID string) {
	workspaces, err := manager.ListWorkspaces()
	if err != nil {
		fmt.Printf("获取工作区列表失败: %v\n", err)
		return
	}

	workspacePath, exists := workspaces[workspaceID]
	if !exists {
		fmt.Printf("工作区 %s 不存在\n", workspaceID)
		return
	}

	fmt.Printf("工作区: %s\n\n", workspacePath)

	// 显示聊天历史
	history, err := manager.GetWorkspaceChatHistory(workspaceID)
	if err != nil {
		fmt.Printf("获取会话历史失败: %v\n", err)
		return
	}

	if len(history) == 0 {
		fmt.Println("没有会话历史")
		return
	}

	fmt.Printf("会话历史 (%d 条):\n", len(history))
	fmt.Println("----------------------------------------")

	for i, h := range history {
		timestamp := time.UnixMilli(h.Timestamp).Format("2006-01-02 15:04:05")
		fmt.Printf("%d. %s\n", i+1, h.Title)
		fmt.Printf("   时间: %s\n", timestamp)
		fmt.Printf("   会话ID: %s\n", h.SessionID)
		fmt.Printf("   上下文: %d 项\n", len(h.Context))
		fmt.Println("----------------------------------------")
	}

	// 显示聊天视图
	views, err := manager.GetWorkspaceChatViews(workspaceID)
	if err == nil && len(views.Views) > 0 {
		fmt.Printf("\n当前打开的聊天视图 (%d 个):\n", len(views.Views))
		for _, v := range views.Views {
			activeMark := ""
			if v.Active {
				activeMark = " [活动]"
			}
			fmt.Printf("  - %s%s\n", v.Title, activeMark)
		}
	}

	// 显示聊天标签
	tabs, err := manager.GetWorkspaceChatTabs(workspaceID)
	if err == nil && len(tabs.Tabs) > 0 {
		fmt.Printf("\n聊天标签 (%d 个):\n", len(tabs.Tabs))
		for _, t := range tabs.Tabs {
			activeMark := ""
			if t.TabID == tabs.ActiveTabID {
				activeMark = " [活动]"
			}
			fmt.Printf("  - %s%s\n", t.Title, activeMark)
		}
	}
}

func exportWorkspace(manager *qoder.QoderSessionManager, workspaceID string) {
	if workspaceID == "all" {
		exportAll(manager)
		return
	}

	workspaces, err := manager.ListWorkspaces()
	if err != nil {
		fmt.Printf("获取工作区列表失败: %v\n", err)
		return
	}

	workspacePath, exists := workspaces[workspaceID]
	if !exists {
		fmt.Printf("工作区 %s 不存在\n", workspaceID)
		return
	}

	backup, err := manager.BackupWorkspace(workspaceID, workspacePath)
	if err != nil {
		fmt.Printf("导出失败: %v\n", err)
		return
	}

	// 导出为可读格式
	exportDir := getDefaultBackupDir()
	timestamp := time.Now().Format("20060102-150405")
	exportPath := filepath.Join(exportDir, fmt.Sprintf("qoder-export-%s.txt", timestamp))

	file, err := os.Create(exportPath)
	if err != nil {
		fmt.Printf("创建导出文件失败: %v\n", err)
		return
	}
	defer file.Close()

	fmt.Fprintf(file, "Qoder 会话导出\n")
	fmt.Fprintf(file, "================\n\n")
	fmt.Fprintf(file, "工作区: %s\n", backup.WorkspacePath)
	fmt.Fprintf(file, "用户ID: %s\n", backup.UserID)
	fmt.Fprintf(file, "导出时间: %s\n\n", backup.BackupTime.Format("2006-01-02 15:04:05"))

	fmt.Fprintf(file, "会话历史 (%d 条):\n", len(backup.ChatHistory))
	fmt.Fprintf(file, "--------------------\n\n")

	for i, h := range backup.ChatHistory {
		timestamp := time.UnixMilli(h.Timestamp).Format("2006-01-02 15:04:05")
		fmt.Fprintf(file, "%d. %s\n", i+1, h.Title)
		fmt.Fprintf(file, "   时间: %s\n", timestamp)
		fmt.Fprintf(file, "   会话ID: %s\n", h.SessionID)

		if len(h.Context) > 0 {
			fmt.Fprintf(file, "   上下文:\n")
			for _, c := range h.Context {
				fmt.Fprintf(file, "     - %s: %s\n", c.Type, c.Name)
			}
		}
		fmt.Fprintln(file)
	}

	fmt.Printf("导出已保存到: %s\n", exportPath)
}

func exportAll(manager *qoder.QoderSessionManager) {
	workspaces, err := manager.ListWorkspaces()
	if err != nil {
		fmt.Printf("获取工作区列表失败: %v\n", err)
		return
	}

	exportDir := getDefaultBackupDir()
	timestamp := time.Now().Format("20060102-150405")
	exportPath := filepath.Join(exportDir, fmt.Sprintf("qoder-export-all-%s.txt", timestamp))

	file, err := os.Create(exportPath)
	if err != nil {
		fmt.Printf("创建导出文件失败: %v\n", err)
		return
	}
	defer file.Close()

	fmt.Fprintf(file, "Qoder 所有工作区会话导出\n")
	fmt.Fprintf(file, "========================\n\n")
	fmt.Fprintf(file, "导出时间: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	for workspaceID, workspacePath := range workspaces {
		backup, err := manager.BackupWorkspace(workspaceID, workspacePath)
		if err != nil {
			fmt.Fprintf(file, "导出 %s 失败: %v\n\n", workspacePath, err)
			continue
		}

		fmt.Fprintf(file, "工作区: %s\n", backup.WorkspacePath)
		fmt.Fprintf(file, "会话数: %d\n\n", len(backup.ChatHistory))

		for i, h := range backup.ChatHistory {
			timestamp := time.UnixMilli(h.Timestamp).Format("2006-01-02 15:04:05")
			fmt.Fprintf(file, "  %d. %s\n", i+1, h.Title)
			fmt.Fprintf(file, "     时间: %s\n", timestamp)
			fmt.Fprintf(file, "     会话ID: %s\n", h.SessionID)
			fmt.Fprintln(file)
		}
		fmt.Fprintln(file)
	}

	fmt.Printf("导出已保存到: %s\n", exportPath)
}

func listBackups(manager *qoder.QoderSessionManager, backupDir string) {
	backups, err := manager.ListBackups(backupDir)
	if err != nil {
		fmt.Printf("获取备份列表失败: %v\n", err)
		return
	}

	if len(backups) == 0 {
		fmt.Printf("目录 %s 中没有找到备份\n", backupDir)
		return
	}

	fmt.Printf("找到 %d 个备份:\n\n", len(backups))

	for i, backup := range backups {
		fmt.Printf("%d. 工作区: %s\n", i+1, backup.WorkspacePath)
		fmt.Printf("   备份时间: %s\n", backup.BackupTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("   用户ID: %s\n", backup.UserID)
		fmt.Printf("   会话数: %d\n", len(backup.ChatHistory))
		fmt.Println("----------------------------------------")
	}
}

func getDefaultBackupDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "Documents", "qoder-backups")
}
