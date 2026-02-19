@echo off
set "APP_NAME=go-connect-tool.exe"

echo 正在尝试关闭程序 %APP_NAME%...
taskkill /F /IM %APP_NAME% >nul 2>&1
if %errorlevel% equ 0 (
    echo 程序已成功关闭。
) else (
    echo 未发现正在运行的程序，或者关闭失败。
)
pause
