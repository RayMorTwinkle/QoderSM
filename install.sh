#!/bin/bash

# Qoder Session Manager 安装脚本

INSTALL_DIR="/usr/local/bin"
EXEC_PATH="/Users/lucky/Documents/WorkSapce/QoderSM/build/bin/qoder-sm"

echo "正在安装 Qoder Session Manager..."

# 检查文件是否存在
if [ ! -f "$EXEC_PATH" ]; then
    echo "错误: 找不到编译后的文件"
    echo "请先运行: go build -o build/bin/qoder-sm ."
    exit 1
fi

# 创建符号链接
sudo ln -sf "$EXEC_PATH" "$INSTALL_DIR/qoder-sm"

if [ $? -eq 0 ]; then
    echo "安装成功！"
    echo "现在可以在任何位置使用 'qoder-sm' 命令"
    echo ""
    echo "试试运行: qoder-sm list"
else
    echo "安装失败"
    exit 1
fi
