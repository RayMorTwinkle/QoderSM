#!/bin/bash

# Qoder 会话自动备份脚本
# 用于在切换用户前自动备份会话

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
QODER_SM="$SCRIPT_DIR/build/bin/qoder-sm"
BACKUP_DIR="$HOME/Documents/qoder-backups"

# 创建备份目录
mkdir -p "$BACKUP_DIR"

echo "=== Qoder 会话备份 ==="
echo "开始时间: $(date)"
echo ""

# 备份所有工作区
if [ -f "$QODER_SM" ]; then
    "$QODER_SM" backup all
else
    echo "错误: 找不到 qoder-sm 命令"
    echo "路径: $QODER_SM"
    exit 1
fi

echo ""
echo "备份完成！"
echo "备份位置: $BACKUP_DIR"
echo ""

# 列出最近的备份
echo "最近的备份:"
ls -lt "$BACKUP_DIR"/*.json 2>/dev/null | head -5
