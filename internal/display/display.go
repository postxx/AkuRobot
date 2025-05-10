package display

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"aku-web/internal/config"
)

var (
	displayMutex sync.Mutex
	currentCmd   *exec.Cmd
)

// DisplayConfig 显示配置
type DisplayConfig struct {
	TempDir string // 临时文件存储目录
}

// Manager 显示管理器
type Manager struct {
	config DisplayConfig
}

// NewManager 创建新的显示管理器
func NewManager(config DisplayConfig) (*Manager, error) {
	// 确保临时目录存在
	if err := os.MkdirAll(config.TempDir, 0755); err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %v", err)
	}

	return &Manager{
		config: config,
	}, nil
}

// ShowText 显示文字
func (m *Manager) ShowText(text string, fontSize int, color string, hAlign, vAlign int) error {
	displayMutex.Lock()
	defer displayMutex.Unlock()

	// 停止当前正在运行的显示进程
	m.stopCurrent()

	cmd := exec.Command(config.ShowTextPath, text, fmt.Sprint(fontSize), color, fmt.Sprint(hAlign), fmt.Sprint(vAlign))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动显示文字失败: %v", err)
	}

	currentCmd = cmd
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("显示文字进程退出: %v", err)
		}
	}()

	return nil
}

// ShowImage 显示图片
func (m *Manager) ShowImage(imagePath string) error {
	displayMutex.Lock()
	defer displayMutex.Unlock()

	// 停止当前正在运行的显示进程
	m.stopCurrent()

	cmd := exec.Command(config.ShowImgPath, imagePath)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动显示图片失败: %v", err)
	}

	currentCmd = cmd
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("显示图片进程退出: %v", err)
		}
	}()

	return nil
}

// ShowGif 显示动图
func (m *Manager) ShowGif(directory string, delayMs int, loop_once bool) error {
	displayMutex.Lock()
	defer displayMutex.Unlock()

	// 停止当前正在运行的显示进程
	m.stopCurrent()

	args := []string{"-d", fmt.Sprint(delayMs)}
	if loop_once {
		args = append(args, "-l")
	}
	args = append(args, directory)

	cmd := exec.Command(config.ShowGifPath, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动显示动图失败: %v", err)
	}

	currentCmd = cmd
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("显示动图进程退出: %v", err)
		}
	}()

	return nil
}

// SaveUploadedFile 保存上传的文件到临时目录
func (m *Manager) SaveUploadedFile(file io.Reader, filename string) (string, error) {
	// 生成唯一的文件名
	ext := filepath.Ext(filename)
	uniqueName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	fullPath := filepath.Join(m.config.TempDir, uniqueName)

	// 创建目标文件
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %v", err)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("保存文件失败: %v", err)
	}

	return fullPath, nil
}

// CleanupOldFiles 清理旧的临时文件
func (m *Manager) CleanupOldFiles(maxAge time.Duration) error {
	now := time.Now()
	return filepath.Walk(m.config.TempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && now.Sub(info.ModTime()) > maxAge {
			if err := os.Remove(path); err != nil {
				log.Printf("删除旧文件失败 %s: %v", path, err)
			}
		}
		return nil
	})
}

// stopCurrent 停止当前正在运行的显示进程
func (m *Manager) stopCurrent() {
	if currentCmd != nil && currentCmd.Process != nil {
		// 先检查进程是否还在运行
		if err := currentCmd.Process.Signal(syscall.Signal(0)); err != nil {
			// 进程已经不存在或无法访问，直接清理
			currentCmd = nil
			return
		}

		// 进程还在运行，尝试终止它
		if err := currentCmd.Process.Kill(); err != nil {
			log.Printf("停止当前显示进程失败: %v", err)
		}
		currentCmd = nil
	}
}

// GetConfig 获取显示管理器配置
func (m *Manager) GetConfig() DisplayConfig {
	return m.config
}
