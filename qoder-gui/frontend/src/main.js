import { GetWorkspaces, GetWorkspaceChats, BackupAll, GetBackups, RestoreBackup, ExportSessions, DeleteBackup, GetUserID } from '../wailsjs/go/main/App.js'

// 全局状态
const state = {
    workspaces: [],
    selectedWorkspace: null,
    chats: [],
    backups: [],
    currentView: 'welcome',
    status: null,
    userId: null
}

// 日志辅助函数
function log(msg) {
    console.log(`[QoderGUI] ${msg}`)
}

// 直接暴露全局函数，不使用闭包
window.qoderRestoreBackup = async function(path) {
    log(`Restore backup called: ${path}`)
    showStatus('正在恢复...', 'info')
    try {
        const result = await RestoreBackup(path)
        if (result.success) {
            showStatus(result.message, 'success')
        } else {
            showStatus(result.error || '恢复失败', 'error')
        }
    } catch (e) {
        log(`Restore error: ${e}`)
        showStatus('恢复失败', 'error')
    }
}

window.qoderDeleteBackup = async function(path) {
    log(`Delete backup called: ${path}`)
    try {
        const success = await DeleteBackup(path)
        if (success) {
            showStatus('备份已删除', 'success')
            await loadBackupsInternal()
        } else {
            showStatus('删除失败', 'error')
        }
    } catch (e) {
        log(`Delete error: ${e}`)
        showStatus('删除失败', 'error')
    }
}

window.qoderSelectWorkspace = function(id) {
    log(`Select workspace: ${id}`)
    loadChatsInternal(id)
}

window.qoderBackupAll = async function() {
    log('Backup all')
    showStatus('正在备份...', 'info')
    try {
        const result = await BackupAll()
        if (result.success) {
            showStatus(result.message, 'success')
            await loadBackupsInternal()
        } else {
            showStatus(result.message || '备份失败', 'error')
        }
    } catch (e) {
        showStatus('备份失败', 'error')
    }
}

window.qoderLoadBackups = async function() {
    log('Load backups')
    await loadBackupsInternal()
}

window.qoderExportSessions = async function() {
    log('Export sessions')
    showStatus('正在导出...', 'info')
    try {
        const path = await ExportSessions()
        showStatus(`已导出到: ${path}`, 'success')
    } catch (e) {
        showStatus('导出失败', 'error')
    }
}

window.qoderRefresh = async function() {
    log('Refresh')
    state.selectedWorkspace = null
    state.currentView = 'welcome'
    await loadWorkspacesInternal()
    showStatus('已刷新', 'success')
}

// 内部加载函数
async function loadWorkspacesInternal() {
    try {
        state.workspaces = await GetWorkspaces()
        log(`Loaded ${state.workspaces.length} workspaces`)
        render()
    } catch (e) {
        log(`Load workspaces error: ${e}`)
        showStatus('加载工作区失败', 'error')
    }
}

async function loadChatsInternal(workspaceId) {
    try {
        state.chats = await GetWorkspaceChats(workspaceId)
        state.selectedWorkspace = workspaceId
        state.currentView = 'chats'
        render()
    } catch (e) {
        log(`Load chats error: ${e}`)
        showStatus('加载会话失败', 'error')
    }
}

async function loadBackupsInternal() {
    try {
        state.backups = await GetBackups()
        state.currentView = 'backups'
        render()
    } catch (e) {
        log(`Load backups error: ${e}`)
        showStatus('加载备份失败', 'error')
    }
}

// 辅助函数
function escapeHtml(text) {
    const div = document.createElement('div')
    div.textContent = text
    return div.innerHTML
}

function formatSize(bytes) {
    if (bytes < 1024) return bytes + ' B'
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function formatTime(timestamp) {
    const date = new Date(timestamp)
    return date.toLocaleString('zh-CN', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
    })
}

function showStatus(message, type = 'info') {
    state.status = { message, type }
    render()
    setTimeout(() => {
        state.status = null
        render()
    }, 3000)
}

// 渲染函数
function render() {
    const app = document.getElementById('app')
    if (!app) {
        log('App element not found!')
        return
    }

    const workspaceItems = state.workspaces.map(ws => `
        <div class="workspace-item ${state.selectedWorkspace === ws.id ? 'active' : ''}"
             onclick="window.qoderSelectWorkspace('${escapeHtml(ws.id)}')">
            <div class="name">${escapeHtml(ws.name)}</div>
            <div class="count">${ws.chatCount} 个会话</div>
        </div>
    `).join('')

    const contentHtml = renderContent()

    app.innerHTML = `
        <div class="main-container">
            <div class="sidebar">
                <div class="sidebar-header">
                    <h1>⬡ Qoder 会话管理器</h1>
                </div>
                <div class="user-info" id="user-info">${state.userId ? `用户: ${escapeHtml(state.userId)}` : '加载中...'}</div>
                <div class="workspace-list">
                    ${workspaceItems}
                </div>
            </div>
            <div class="main-content">
                <div class="toolbar">
                    <button class="btn btn-primary" onclick="window.qoderBackupAll()">备份所有会话</button>
                    <button class="btn" onclick="window.qoderLoadBackups()">查看备份</button>
                    <button class="btn" onclick="window.qoderExportSessions()">导出会话</button>
                    <button class="btn" onclick="window.qoderRefresh()">刷新</button>
                </div>
                <div class="content">
                    ${contentHtml}
                </div>
            </div>
        </div>
        ${state.status ? `
            <div class="status-bar ${state.status.type}">
                <span>${state.status.message}</span>
            </div>
        ` : ''}
    `
}

function renderContent() {
    if (state.currentView === 'welcome') {
        return `
            <div class="welcome">
                <h2>欢迎使用 Qoder 会话管理器</h2>
                <p>选择左侧工作区查看会话，或使用上方按钮进行操作</p>
                <div class="welcome-cards">
                    <div class="welcome-card">
                        <div class="icon">💾</div>
                        <h3>快速备份</h3>
                        <p>一键备份所有会话</p>
                    </div>
                    <div class="welcome-card">
                        <div class="icon">🔄</div>
                        <h3>轻松恢复</h3>
                        <p>切换用户后恢复</p>
                    </div>
                    <div class="welcome-card">
                        <div class="icon">📊</div>
                        <h3>查看历史</h3>
                        <p>浏览所有会话记录</p>
                    </div>
                </div>
            </div>
        `
    }

    if (state.currentView === 'chats') {
        if (state.chats.length === 0) {
            return `
                <div class="empty">
                    <div class="empty-icon">💬</div>
                    <p>该工作区暂无会话记录</p>
                </div>
            `
        }
        return `
            <div class="section-title">会话历史 (${state.chats.length})</div>
            <div class="chat-list">
                ${state.chats.map(chat => `
                    <div class="chat-item">
                        <div class="title">${escapeHtml(chat.title || '无标题')}</div>
                        <div class="meta">
                            <span>🕐 ${chat.timeStr}</span>
                            <span>📋 ${chat.id.substring(0, 8)}</span>
                        </div>
                    </div>
                `).join('')}
            </div>
        `
    }

    if (state.currentView === 'backups') {
        if (state.backups.length === 0) {
            return `
                <div class="empty">
                    <div class="empty-icon">📦</div>
                    <p>暂无备份文件</p>
                    <p class="small">点击"备份所有会话"创建第一个备份</p>
                </div>
            `
        }
        return `
            <div class="section-title">备份列表 (${state.backups.length})</div>
            <div class="backup-list">
                ${state.backups.map((backup) => `
                    <div class="backup-item">
                        <div class="info">
                            <div class="name">${escapeHtml(backup.name)}</div>
                            <div class="meta">
                                ${formatTime(backup.modTime * 1000)} · ${formatSize(backup.size)}
                            </div>
                        </div>
                        <div class="actions">
                            <button class="btn btn-primary" onclick="window.qoderRestoreBackup('${escapeHtml(backup.path).replace(/'/g, "\\'")}')">恢复</button>
                            <button class="btn btn-danger" onclick="window.qoderDeleteBackup('${escapeHtml(backup.path).replace(/'/g, "\\'")}')">删除</button>
                        </div>
                    </div>
                `).join('')}
            </div>
        `
    }

    return ''
}

// 初始化
async function init() {
    log('=== 应用启动 ===')

    // 加载用户信息
    try {
        state.userId = await GetUserID()
        log(`User ID: ${state.userId}`)
    } catch (e) {
        log(`Load user error: ${e}`)
        state.userId = '加载失败'
    }

    // 加载工作区
    await loadWorkspacesInternal()

    log('=== 应用初始化完成 ===')
}

// 启动应用
log('脚本已加载，等待DOM准备...')
document.addEventListener('DOMContentLoaded', () => {
    log('DOM已准备，启动应用')
    init()
})
