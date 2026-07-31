.PHONY: all build clean install test web web-run

# 默认构建CLI
all: cli web

# CLI 构建
cli:
	go build -o build/bin/qoder-sm .

# Web 构建
web:
	go build -o build/bin/qoder-web cmd/web/main.go

# 运行Web版本
web-run: web
	./build/bin/qoder-web

# 安装Web版本（创建.app包）
web-install: web
	@echo "创建Web应用包..."
	@rm -rf build/QoderSessionManager.app
	@mkdir -p build/QoderSessionManager.app/Contents/MacOS
	@mkdir -p build/QoderSessionManager.app/Contents/Resources/web
	@echo '<?xml version="1.0" encoding="UTF-8"?>' > build/QoderSessionManager.app/Contents/Info.plist
	@echo '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' >> build/QoderSessionManager.app/Contents/Info.plist
	@echo '<plist version="1.0">' >> build/QoderSessionManager.app/Contents/Info.plist
	@echo '<dict>' >> build/QoderSessionManager.app/Contents/Info.plist
	@echo '  <key>CFBundleExecutable</key>' >> build/QoderSessionManager.app/Contents/Info.plist
	@echo '  <string>qoder-web</string>' >> build/QoderSessionManager.app/Contents/Info.plist
	@echo '  <key>CFBundleIdentifier</key>' >> build/QoderSessionManager.app/Contents/Info.plist
	@echo '  <string>com.qoder.sessionmanager</string>' >> build/QoderSessionManager.app/Contents/Info.plist
	@echo '  <key>CFBundleName</key>' >> build/QoderSessionManager.app/Contents/Info.plist
	@echo '  <string>Qoder Session Manager</string>' >> build/QoderSessionManager.app/Contents/Info.plist
	@echo '  <key>CFBundleVersion</key>' >> build/QoderSessionManager.app/Contents/Info.plist
	@echo '  <string>1.0.0</string>' >> build/QoderSessionManager.app/Contents/Info.plist
	@echo '  <key>NSHighResolutionCapable</key>' >> build/QoderSessionManager.app/Contents/Info.plist
	@echo '  <true/>' >> build/QoderSessionManager.app/Contents/Info.plist
	@echo '</dict>' >> build/QoderSessionManager.app/Contents/Info.plist
	@echo '</plist>' >> build/QoderSessionManager.app/Contents/Info.plist
	@cp build/bin/qoder-web build/QoderSessionManager.app/Contents/MacOS/
	@cp web/index.html build/QoderSessionManager.app/Contents/Resources/web/
	@echo "创建完成: build/QoderSessionManager.app"

# 安装到应用目录
install: web-install
	@if [ -d "/Applications/QoderSessionManager.app" ]; then \
		rm -rf /Applications/QoderSessionManager.app; \
	fi
	@cp -R build/QoderSessionManager.app /Applications/
	@echo "已安装到 /Applications/QoderSessionManager.app"
	@echo "从启动台或应用文件夹启动 Qoder Session Manager"

clean:
	rm -rf build

test: cli
	./build/bin/qoder-sm list

# 获取依赖
deps:
	go mod tidy
