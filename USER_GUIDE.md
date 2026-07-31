# Qoder 会话管理器 - 快速使用指南

## 解决的问题

当你在Qoder中切换用户账号时，之前的聊天会话会丢失。这个工具可以帮你：
- 在切换用户前备份会话
- 在切换用户后恢复会话
- 随时导出和查看会话历史

## 快速开始

### 1. 安装

```bash
cd /Users/lucky/Documents/WorkSapce/QoderSM
make install
```

### 2. 查看当前会话

```bash
qoder-sm list
```

这会显示所有工作区和它们的会话数量。

### 3. 备份所有会话（在切换用户前）

```bash
qoder-sm backup all
```

### 4. 切换Qoder用户

在Qoder中切换你的账号...

### 5. 恢复会话（切换用户后）

```bash
# 查看最近的备份
ls -lt ~/Documents/qoder-backups/

# 恢复最新的备份
qoder-sm restore ~/Documents/qoder-backups/qoder-backup-all-[最新日期].json
```

### 6. 重启Qoder

恢复后需要重启Qoder才能看到会话。

## 常用命令

```bash
# 列出所有工作区
qoder-sm list

# 查看特定工作区的会话详情
qoder-sm show <workspace-id>

# 备份所有工作区
qoder-sm backup all

# 备份特定工作区
qoder-sm backup <workspace-id>

# 恢复备份
qoder-sm restore <备份文件路径>

# 导出会话为文本格式
qoder-sm export all

# 列出所有备份
qoder-sm list-backups
```

## 工作流程示例

### 场景1：日常使用中的会话备份

```bash
# 每天开始工作前备份
qoder-sm backup all

# 工作结束后再次备份
qoder-sm backup all
```

### 场景2：在多个Qoder账号间切换

```bash
# 1. 当前是账号A，备份会话
qoder-sm backup all

# 2. 在Qoder中切换到账号B

# 3. 恢复之前的会话
qoder-sm restore ~/Documents/qoder-backups/qoder-backup-all-[最新].json

# 4. 重启Qoder
```

### 场景3：定期自动备份（使用cron）

```bash
# 编辑crontab
crontab -e

# 添加每天晚上8点自动备份
0 20 * * * /Users/lucky/Documents/WorkSapce/QoderSM/backup.sh >> ~/qoder-backup.log 2>&1
```

## 数据说明

### 会话存储位置

- **会话历史**: `~/Library/Application Support/Qoder/User/globalStorage/state.vscdb`
- **工作区配置**: `~/Library/Application Support/Qoder/User/workspaceStorage/[workspace-id]/state.vscdb`
- **用户ID**: `~/Library/Application Support/Qoder/.auth/id`

### 备份文件内容

备份文件包含：
- 会话历史（所有对话记录）
- 聊天视图（打开的聊天窗口）
- 聊天标签（聊天标签页）
- 工作区信息
- 用户ID
- 备份时间

## 故障排除

### 问题：恢复后看不到会话

**解决方案**：
1. 确保已经重启Qoder
2. 检查备份文件是否包含会话数据
3. 运行 `qoder-sm show <workspace-id>` 检查数据

### 问题：备份失败

**解决方案**：
1. 确保Qoder没有运行（或者确保数据库文件没有被锁定）
2. 检查文件权限

### 问题：编译失败

**解决方案**：
```bash
# 安装依赖
go mod tidy

# 如果缺少sqlite3开发包
brew install sqlite3
```

## 高级用法

### 只恢复特定会话

1. 先导出查看备份内容：
```bash
qoder-sm export all
```

2. 手动编辑备份文件，只保留需要的会话

3. 恢复编辑后的备份

### 合并多个用户的会话

这需要手动操作：
1. 分别备份不同用户的会话
2. 手动合并JSON文件
3. 恢复合并后的文件

## 反馈和改进

如果你有任何问题或建议，请反馈给开发者。
