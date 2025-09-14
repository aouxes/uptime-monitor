#!/bin/bash

# Скрипт для сборки приложения

set -e

echo "🚀 Сборка Uptime Monitor..."

# Проверка наличия Go
if ! command -v go &> /dev/null; then
    echo "❌ Go не установлен. Установите Go 1.21+ и попробуйте снова."
    exit 1
fi

# Проверка версии Go
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.21"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "❌ Требуется Go версии $REQUIRED_VERSION или выше. Текущая версия: $GO_VERSION"
    exit 1
fi

echo "✅ Go версии $GO_VERSION найден"

# Создание директории для билдов
mkdir -p build

# Сборка для разных платформ
echo "📦 Сборка для Linux..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/uptime-monitor-linux cmd/server/main.go

echo "📦 Сборка для Windows..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o build/uptime-monitor.exe cmd/server/main.go

echo "📦 Сборка для macOS..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o build/uptime-monitor-macos cmd/server/main.go

# Сборка для текущей платформы
echo "📦 Сборка для текущей платформы..."
go build -ldflags="-s -w" -o build/uptime-monitor cmd/server/main.go

# Копирование статических файлов
echo "📁 Копирование статических файлов..."
cp -r web build/
cp -r migrations build/

# Создание архива
echo "🗜️ Создание архива..."
cd build
tar -czf ../uptime-monitor-linux.tar.gz uptime-monitor-linux web migrations
zip -r ../uptime-monitor-windows.zip uptime-monitor.exe web migrations
tar -czf ../uptime-monitor-macos.tar.gz uptime-monitor-macos web migrations
cd ..

echo "✅ Сборка завершена!"
echo "📁 Файлы находятся в директории build/"
echo "📦 Архивы созданы:"
echo "   - uptime-monitor-linux.tar.gz"
echo "   - uptime-monitor-windows.zip"
echo "   - uptime-monitor-macos.tar.gz"
