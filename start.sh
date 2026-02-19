#!/bin/bash
APP_NAME="go-connect-tool"

echo "正在构建程序..."
go build -o $APP_NAME main.go
if [ $? -ne 0 ]; then
    echo "构建失败，请检查 Go 环境。"
    exit 1
fi

echo "正在后台启动程序..."
nohup ./$APP_NAME > output.log 2>&1 &
echo "程序已在后台启动。您可以访问 http://localhost:8080"
echo "日志输出在 output.log"
