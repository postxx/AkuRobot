@echo off
REM 设置中文编码
chcp 65001
echo.

REM 设置Go的完整路径
set GO_PATH=C:\Program Files\Go\bin\go.exe

REM 检查Go环境
if not exist "%GO_PATH%" (
    echo Error: 未找到Go程序，请检查路径：%GO_PATH%
    pause
    exit /b 1
)

REM 设置交叉编译的环境变量
set GOOS=linux
set GOARCH=arm
set GOARM=7
set CGO_ENABLED=0

echo [INFO] 正在编译 ARM 版本...

REM 编译到当前目录
"%GO_PATH%" build -o aku-web

if %ERRORLEVEL% EQU 0 (
    echo [INFO] ARM 版本编译成功！
) else (
    echo [ERROR] ARM 版本编译失败！
    exit /b 1
)

REM 检查static目录是否存在
if not exist static (
    echo [WARN] static目录不存在，跳过打包静态文件
    goto :skip_tar
)

echo [INFO] 正在打包静态文件...
if exist static (
    tar -czf static.tar.gz static
    if %ERRORLEVEL% EQU 0 (
        echo [INFO] 静态文件打包成功！
    ) else (
        echo [ERROR] 静态文件打包失败！
    )
)

:skip_tar
echo [INFO] 编译完成！生成的文件：
dir /B aku-web static.tar.gz 2>nul 