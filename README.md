# Qoder 会话管理器

解决Qoder IDE切换用户时会话丢失的问题，提供图形化界面管理会话。
<img width="2000" height="1456" alt="7640411f-fb78-4123-97f5-14d358e3942a" src="https://github.com/user-attachments/assets/013d84a5-6833-4d10-9f7a-f671c8ee4284" />

## 功能特性

- 🖥️ **Web图形界面** - 美观的Web界面查看和管理会话
- 💾 **快速备份** - 一键备份所有工作区的聊天会话
- 🔄 **轻松恢复** - 快速恢复之前的会话记录
- 📊 **会话浏览** - 查看所有工作区的会话历史
- 📤 **导出功能** - 将会话导出为Markdown格式

## 快速开始

### 安装

```bash
cd /Users/lucky/Documents/WorkSapce/QoderSM
make deps
make install
```

安装后，从**应用程序**文件夹启动"Qoder Session Manager"

### 使用

1. 启动应用后，会自动在浏览器中打开Web界面
2. 左侧显示所有工作区列表，点击可查看会话
3. 点击"备份所有会话"按钮进行备份
4. 切换Qoder用户后，点击"查看备份"→"恢复"即可恢复会话
5. 重启Qoder查看恢复的会话

## 命令行工具

```bash
# 列出所有工作区
build/bin/qoder-sm list

# 备份所有会话
build/bin/qoder-sm backup all

# 恢复会话
build/bin/qoder-sm restore ~/Documents/qoder-backups/qoder-backup-all-xxx.json
```

## 在Qoder中的使用建议

### 切换用户前

1. 打开Qoder Session Manager
2. 点击"备份所有会话"
3. 在Qoder中切换账号

### 切换用户后

1. 打开Qoder Session Manager
2. 点击"查看备份"
3. 选择备份文件，点击"恢复"
4. 重启Qoder

## 文件位置

- 应用: `/Applications/QoderSessionManager.app`
- 备份: `~/Documents/qoder-backups/`
- 源码: `/Users/lucky/Documents/WorkSapce/QoderSM/`
