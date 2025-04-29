package main

import (
	"aku-web/internal/config"
	"aku-web/internal/server"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

const banner = `

________  ___  __    ___  ___          ___       __   _______   ________     
|\   __  \|\  \|\  \ |\  \|\  \        |\  \     |\  \|\  ___ \ |\   __  \    
\ \  \|\  \ \  \/  /|\ \  \\\  \       \ \  \    \ \  \ \   __/|\ \  \|\ /_   
 \ \   __  \ \   ___  \ \  \\\  \       \ \  \  __\ \  \ \  \_|/_\ \   __  \  
  \ \  \ \  \ \  \\ \  \ \  \\\  \       \ \  \|\__\_\  \ \  \_|\ \ \  \|\  \ 
   \ \__\ \__\ \__\\ \__\ \_______\       \ \____________\ \_______\ \_______\
    \|__|\|__|\|__| \|__|\|_______|        \|____________|\|_______|\|_______|

`

const (
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorReset   = "\033[0m"
)

func printColorized(color, format string, a ...interface{}) {
	fmt.Printf(color+format+colorReset+"\n", a...)
}

func showProgress(prefix string, duration time.Duration) {
	steps := 20
	delay := duration / time.Duration(steps)

	fmt.Printf("%s [", prefix)
	for i := 0; i < steps; i++ {
		time.Sleep(delay)
		fmt.Printf(colorGreen + "=" + colorReset)
		if i < steps-1 {
			fmt.Printf(">")
		}
		fmt.Printf("\b")
	}
	fmt.Printf("] %s完成%s\n", colorGreen, colorReset)
}

func main() {
	// 清屏
	fmt.Print("\033[H\033[2J")

	// 打印炫酷的启动画面
	fmt.Print(colorCyan + banner + colorReset)
	time.Sleep(1 * time.Second)

	// 显示初始化进度
	printColorized(colorMagenta, "\n=== 系统初始化 ===")
	showProgress("初始化日志系统    ", 500*time.Millisecond)

	// 设置日志输出到控制台和文件
	logFile, err := os.OpenFile("aku-web.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("无法打开日志文件: %v，将只输出到控制台", err)
	} else {
		defer logFile.Close()
		log.SetOutput(os.Stdout)
		if logFile != nil {
			mw := io.MultiWriter(os.Stdout, logFile)
			log.SetOutput(mw)
		}
	}

	showProgress("加载系统配置      ", 300*time.Millisecond)
	showProgress("初始化服务模块    ", 400*time.Millisecond)
	showProgress("检查系统依赖      ", 600*time.Millisecond)

	// 打印系统信息
	printColorized(colorMagenta, "\n=== Aku Web 启动信息 ===")
	printColorized(colorGreen, "✓ 系统信息:")
	printColorized(colorYellow, "  • 操作系统: %s/%s", runtime.GOOS, runtime.GOARCH)
	printColorized(colorYellow, "  • CPU核心: %d", runtime.NumCPU())
	printColorized(colorYellow, "  • Go版本: %s", runtime.Version())

	printColorized(colorGreen, "\n✓ 项目信息:")
	printColorized(colorYellow, "  • 开发者: HlameMastar")
	printColorized(colorYellow, "  • 开源地址: https://github.com/jimieguang/AkuRobot")
	printColorized(colorYellow, "  • 版本: v1.2.0")
	printColorized(colorYellow, "  • 构建时间: %s", time.Now().Format("2006-01-02 15:04:05"))

	printColorized(colorGreen, "\n✓ 服务状态:")
	printColorized(colorYellow, "  • HTTP服务: 启动中...")
	printColorized(colorYellow, "  • 监听端口: %s", config.DefaultPort)
	printColorized(colorYellow, "  • 静态目录: %s", config.DefaultDir)

	// 打印分隔线
	printColorized(colorMagenta, "\n"+strings.Repeat("=", 50))

	// 启动HTTP服务器
	if err := server.Start(); err != nil {
		printColorized(colorRed, "✗ 服务器错误: %v", err)
		os.Exit(1)
	}
}
