@echo off
set "APP_NAME=go-connect-tool.exe"

echo 正在构建程序...
go build -o %APP_NAME% main.go
if %errorlevel% neq 0 (
    echo 构建失败，请检查 Go 环境。
    pause
    exit /b %errorlevel%
)

echo 正在启动程序...
start "" %APP_NAME%
echo 程序已启动。您可以访问 http://localhost:8080
pause
