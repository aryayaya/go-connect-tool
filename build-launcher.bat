@echo off
chcp 65001 >nul

echo [1/3] 生成 ICO 图标文件...
go run ./tools/makeico/
if %errorlevel% neq 0 ( echo 图标生成失败。 & pause & exit /b %errorlevel% )

echo [2/3] 嵌入 Windows 资源（rsrc）...
rsrc -ico launcher/icon.ico -o launcher/rsrc.syso
if %errorlevel% neq 0 (
    echo rsrc 未安装，正在安装...
    go install github.com/akavel/rsrc@latest
    rsrc -ico launcher/icon.ico -o launcher/rsrc.syso
)

echo [3/3] 编译托盘启动器...
go build -ldflags "-H windowsgui" -o launcher.exe ./launcher/
if %errorlevel% neq 0 ( echo 编译失败，请检查 Go 环境。 & pause & exit /b %errorlevel% )

echo 完成！launcher.exe 已生成（含自定义图标）。
pause

