#!/bin/bash
APP_NAME="go-connect-tool"

echo "正在尝试关闭程序 $APP_NAME..."
PID=$(pgrep -f "./$APP_NAME")
if [ -z "$PID" ]; then
    echo "未发现正在运行的程序。"
else
    kill $PID
    echo "程序 (PID: $PID) 已成功关闭。"
fi
