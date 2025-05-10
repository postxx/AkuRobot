package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"aku-web/internal/display"
)

var displayManager *display.Manager

// InitDisplayManager 初始化显示管理器
func InitDisplayManager(tempDir string) error {
	var err error
	displayManager, err = display.NewManager(display.DisplayConfig{
		TempDir: tempDir,
	})
	return err
}

// HandleShowText 处理显示文字的请求
func HandleShowText(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Text     string `json:"text"`
		FontSize int    `json:"fontSize"`
		Color    string `json:"color"`
		HAlign   int    `json:"hAlign"`
		VAlign   int    `json:"vAlign"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := displayManager.ShowText(request.Text, request.FontSize, request.Color, request.HAlign, request.VAlign); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleShowImage 处理显示图片的请求
func HandleShowImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析多部分表单
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 限制 10MB
		http.Error(w, "文件太大", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "无法获取上传的文件", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 保存文件
	savedPath, err := displayManager.SaveUploadedFile(file, header.Filename)
	if err != nil {
		http.Error(w, fmt.Sprintf("保存文件失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 显示图片
	if err := displayManager.ShowImage(savedPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// HandleShowGif 处理显示动图的请求
func HandleShowGif(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析多部分表单
	if err := r.ParseMultipartForm(20 << 20); err != nil { // 限制 20MB
		http.Error(w, "文件太大", http.StatusBadRequest)
		return
	}

	// 获取参数
	delayStr := r.FormValue("delay")
	loopStr := r.FormValue("loop")

	delay := 100 // 默认延迟 100ms
	if delayStr != "" {
		if d, err := strconv.Atoi(delayStr); err == nil {
			delay = d
		}
	}

	loop_once := true // 默认单次播放
	if loopStr == "true" {
		loop_once = false
	}

	// 创建临时目录存储帧图片
	tempDir := filepath.Join(displayManager.GetConfig().TempDir, fmt.Sprintf("gif_%d", time.Now().UnixNano()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		http.Error(w, "创建临时目录失败", http.StatusInternalServerError)
		return
	}

	// 获取上传的文件
	files := r.MultipartForm.File["frames"]
	if len(files) == 0 {
		http.Error(w, "没有上传帧图片", http.StatusBadRequest)
		return
	}

	// 保存所有帧图片
	for i, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, fmt.Sprintf("打开文件失败: %v", err), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		framePath := filepath.Join(tempDir, fmt.Sprintf("frame_%04d.bmp", i))
		dst, err := os.Create(framePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("创建帧文件失败: %v", err), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, fmt.Sprintf("保存帧文件失败: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// 显示动图
	if err := displayManager.ShowGif(tempDir, delay, loop_once); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CleanupDisplayFiles 清理旧的显示文件
func CleanupDisplayFiles(maxAge time.Duration) error {
	return displayManager.CleanupOldFiles(maxAge)
}
