package player

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// AudioDuration 存储音频时长信息
type AudioDuration struct {
	Minutes      int
	Seconds      int
	Centiseconds int     // 百分之一秒（0-99）
	TotalSeconds float64 // 总秒数，包含小数部分
	TotalFrames  int     // 总帧数
}

// CacheStatus 缓存状态
type CacheStatus int

const (
	CacheStatusDownloading CacheStatus = iota
	CacheStatusCompleted
	CacheStatusError
)

// AudioCache 管理音频缓存
type AudioCache struct {
	CacheDir     string
	MaxCacheSize int64
	files        map[string]*CacheInfo
	mutex        sync.RWMutex
	downloadMu   sync.RWMutex
}

// CacheInfo 存储缓存文件信息
type CacheInfo struct {
	Path       string
	Size       int64
	LastAccess time.Time
	Status     CacheStatus
	Duration   *AudioDuration
	ReadySize  int64 // 已下载的大小
	mutex      sync.RWMutex
}

// AudioPlayer 音频播放器
type AudioPlayer struct {
	cache       *AudioCache
	currentFile string
	duration    *AudioDuration
	isPlaying   bool
	mutex       sync.RWMutex

	// 播放器控制
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

// 全局播放器实例
var (
	defaultPlayer *AudioPlayer
	once          sync.Once
)

// initDefaultPlayer 初始化默认播放器
func initDefaultPlayer() {
	once.Do(func() {
		// 在系统临时目录下创建缓存目录
		cacheDir := filepath.Join(os.TempDir(), "audio_cache")
		var err error
		defaultPlayer, err = NewAudioPlayer(cacheDir)
		if err != nil {
			log.Printf("初始化默认播放器失败: %v", err)
		}
	})
}

// GetAudioDuration 获取音频时长
func GetAudioDuration(url string) (*AudioDuration, error) {
	initDefaultPlayer()
	if defaultPlayer == nil {
		return nil, fmt.Errorf("播放器初始化失败")
	}
	return defaultPlayer.GetDuration(url)
}

// GetDuration 获取音频时长（使用当前播放器实例）
func (p *AudioPlayer) GetDuration(url string) (*AudioDuration, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 检查是否为常见的不支持格式
	lowerUrl := strings.ToLower(url)
	unsupportedFormats := []string{".wav", ".m4a", ".mp4", ".aac", ".ogg", ".flac", ".wma", ".aiff"}
	for _, format := range unsupportedFormats {
		if strings.HasSuffix(lowerUrl, format) {
			return nil, fmt.Errorf("不支持的音频格式: %s，仅支持MP3格式", format)
		}
	}

	// 初始化播放器（如果尚未初始化）
	if err := p.initPlayer(); err != nil {
		return nil, fmt.Errorf("初始化播放器失败: %v", err)
	}

	// 发送加载命令
	if err := p.sendCommand(fmt.Sprintf("LOAD %s", url)); err != nil {
		return nil, fmt.Errorf("发送加载命令失败: %v", err)
	}
	// 禁止无关信息输出
	p.sendCommand("SILENCE")

	// 读取输出直到文件加载完成
	buf := make([]byte, 1024)
	for {
		log.Printf("读取输出")
		n, err := p.stdout.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("读取输出失败: %v", err)
		}
		output := string(buf[:n])
		if strings.Contains(output, "@S") {
			break
		}
	}

	var sampleRate int
	var totalSamples int64

	// 获取格式信息
	if err := p.sendCommand("FORMAT"); err != nil {
		return nil, fmt.Errorf("获取格式信息失败: %v", err)
	}

	// 读取格式信息
	n, err := p.stdout.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("读取格式信息失败: %v", err)
	}
	output := string(buf[:n])
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "@F") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				sampleRate, _ = strconv.Atoi(parts[1])
				break
			}
		}
	}

	// 获取采样信息
	if err := p.sendCommand("SAMPLE"); err != nil {
		return nil, fmt.Errorf("获取采样信息失败: %v", err)
	}

	// 读取采样信息
	n, err = p.stdout.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("读取采样信息失败: %v", err)
	}
	output = string(buf[:n])
	lines = strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "@S") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				totalSamples, _ = strconv.ParseInt(parts[2], 10, 64)
				break
			}
		}
	}

	// 检查是否获取到所需信息
	if sampleRate == 0 || totalSamples == 0 {
		return nil, fmt.Errorf("无法获取完整的音频信息")
	}

	// 停止当前加载的音频
	p.sendCommand("STOP")

	// 计算总时长（秒）
	totalSeconds := float64(totalSamples) / float64(sampleRate)
	minutes := int(totalSeconds) / 60
	seconds := int(totalSeconds) % 60
	centiseconds := int((totalSeconds - float64(int(totalSeconds))) * 100)

	log.Printf("音频信息: 采样率=%d Hz, 总采样数=%d, 总时长=%.2f秒",
		sampleRate, totalSamples, totalSeconds)

	return &AudioDuration{
		Minutes:      minutes,
		Seconds:      seconds,
		Centiseconds: centiseconds,
		TotalSeconds: totalSeconds,
		TotalFrames:  int(totalSamples), // 使用采样数作为帧数
	}, nil
}

// PlayStream 全局播放流媒体方法
func PlayStream(url string) (*AudioDuration, error) {
	initDefaultPlayer()
	if defaultPlayer == nil {
		return nil, fmt.Errorf("播放器初始化失败")
	}
	return defaultPlayer.PlayStream(url)
}

// PausePlayback 暂停播放
func PausePlayback() error {
	initDefaultPlayer()
	if defaultPlayer == nil {
		return fmt.Errorf("播放器初始化失败")
	}
	return defaultPlayer.Pause()
}

// ResumePlayback 继续播放
func ResumePlayback() error {
	initDefaultPlayer()
	if defaultPlayer == nil {
		return fmt.Errorf("播放器初始化失败")
	}
	return defaultPlayer.Resume()
}

// StopPlayback 停止播放
func StopPlayback() {
	if defaultPlayer != nil {
		defaultPlayer.Stop()
	}
}

// SeekTo 跳转到指定位置
func SeekTo(position float64) error {
	initDefaultPlayer()
	if defaultPlayer == nil {
		return fmt.Errorf("播放器初始化失败")
	}
	log.Printf("SeekTo: 跳转到 %.2f 秒", position)
	return defaultPlayer.SeekTo(position)
}

// NewAudioPlayer 创建新的音频播放器实例
func NewAudioPlayer(cacheDir string) (*AudioPlayer, error) {
	cache := &AudioCache{
		CacheDir:     cacheDir,
		MaxCacheSize: 1024 * 1024 * 1024, // 1GB 缓存上限
		files:        make(map[string]*CacheInfo),
	}

	// 创建缓存目录
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("创建缓存目录失败: %v", err)
	}

	return &AudioPlayer{
		cache: cache,
	}, nil
}

// PlayStream 改进的流媒体播放方法
func (p *AudioPlayer) PlayStream(url string) (*AudioDuration, error) {
	log.Printf("[PlayStream] 开始播放，URL: %s", url)

	// 先获取音频时长信息
	duration, err := p.GetDuration(url)
	if err != nil {
		log.Printf("[PlayStream] 获取音频时长失败: %v", err)
		return nil, err
	}
	p.duration = duration
	log.Printf("[PlayStream] 获取音频时长成功: %.2f秒", duration.TotalSeconds)

	p.mutex.Lock()
	defer p.mutex.Unlock()

	log.Printf("[PlayStream] 开始播放，URL: %s", url)

	// 开始缓存并获取缓存信息
	cacheInfo := p.startCaching(url, duration)
	if cacheInfo == nil {
		return nil, fmt.Errorf("创建缓存失败")
	}

	// 初始化播放器
	if err := p.initPlayer(); err != nil {
		log.Printf("[PlayStream] 初始化播放器失败: %v", err)
		return nil, err
	}

	// 等待足够的数据被缓存（至少1MB或文件大小的10%）
	const minBufferSize = 1024 * 1024 // 1MB
	for {
		cacheInfo.mutex.RLock()
		readySize := cacheInfo.ReadySize
		cacheInfo.mutex.RUnlock()

		if readySize >= minBufferSize || (cacheInfo.Size > 0 && readySize >= cacheInfo.Size/10) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// 开始播放
	log.Printf("开始播放: %s", cacheInfo.Path)
	if err := p.sendCommand(fmt.Sprintf("LOAD %s", cacheInfo.Path)); err != nil {
		return nil, fmt.Errorf("加载音频失败: %v", err)
	}

	p.currentFile = url
	p.isPlaying = true

	return duration, nil
}

// startCaching 开始缓存音频文件
func (p *AudioPlayer) startCaching(url string, duration *AudioDuration) *CacheInfo {
	p.cache.downloadMu.Lock()
	defer p.cache.downloadMu.Unlock()

	// 检查是否已经在下载或已缓存
	if cached := p.cache.GetCachedFile(url); cached != nil {
		return cached // 返回现有缓存
	}

	// 在缓存目录中创建临时文件
	tmpFile, err := os.CreateTemp(p.cache.CacheDir, "audio_*")
	if err != nil {
		log.Printf("创建临时文件失败: %v", err)
		return nil
	}

	// 创建新的缓存信息
	cacheInfo := &CacheInfo{
		Path:       tmpFile.Name(),
		LastAccess: time.Now(),
		Status:     CacheStatusDownloading,
		Duration:   duration,
		ReadySize:  0,
	}

	// 添加到缓存管理
	p.cache.mutex.Lock()
	p.cache.files[url] = cacheInfo
	p.cache.mutex.Unlock()

	// 启动下载协程
	go p.downloadAndCache(url, tmpFile, cacheInfo)

	return cacheInfo
}

// downloadAndCache 下载并缓存音频文件
func (p *AudioPlayer) downloadAndCache(url string, tmpFile *os.File, info *CacheInfo) {
	defer tmpFile.Close()

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 获取文件大小
	resp, err := client.Head(url)
	if err != nil {
		log.Printf("获取文件信息失败: %v", err)
		info.Status = CacheStatusError
		return
	}
	fileSize := resp.ContentLength

	// 使用512KB的分片大小，适合音频流
	const chunkSize = 512 * 1024
	chunks := (fileSize + chunkSize - 1) / chunkSize

	// 预分配文件大小
	if err := tmpFile.Truncate(fileSize); err != nil {
		log.Printf("预分配文件大小失败: %v", err)
		info.Status = CacheStatusError
		return
	}

	// 按顺序下载分片，确保流式播放的连续性
	for i := int64(0); i < chunks; i++ {
		start := i * chunkSize
		end := start + chunkSize - 1
		if end > fileSize {
			end = fileSize - 1
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("创建请求失败: %v", err)
			continue
		}
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("下载分片失败: %v", err)
			continue
		}

		if _, err := tmpFile.Seek(start, 0); err != nil {
			resp.Body.Close()
			log.Printf("设置文件位置失败: %v", err)
			continue
		}

		written, err := io.Copy(tmpFile, resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("写入分片失败: %v", err)
			continue
		}

		// 更新已下载大小
		info.mutex.Lock()
		info.ReadySize = start + written
		info.mutex.Unlock()
	}

	info.Size = fileSize
	info.Status = CacheStatusCompleted
}

// SeekTo 改进的跳转方法
func (p *AudioPlayer) SeekTo(position float64) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	log.Printf("[SeekTo] 尝试跳转到 %.2f 秒", position)

	if !p.isPlaying {
		return fmt.Errorf("播放器未在播放状态")
	}

	cached := p.cache.GetCachedFile(p.currentFile)
	if cached == nil {
		return fmt.Errorf("未找到缓存文件")
	}

	// 检查跳转位置是否有效
	if p.duration != nil && position > p.duration.TotalSeconds {
		return fmt.Errorf("跳转位置 (%.2f) 超出音频总长度 (%.2f)", position, p.duration.TotalSeconds)
	}

	// 等待缓存追赶到目标位置
	maxWaitTime := 5 * time.Second
	startWait := time.Now()

	for {
		cached.mutex.RLock()
		readySize := cached.ReadySize
		totalSize := cached.Size
		status := cached.Status
		cached.mutex.RUnlock()

		// 如果已经完成下载，或者已缓存的数据足够，就可以进行跳转
		estimatedRequiredSize := int64(0)
		if p.duration != nil && totalSize > 0 {
			estimatedRequiredSize = int64(float64(totalSize) * (position / p.duration.TotalSeconds))
		}
		log.Printf("estimatedRequiredSize: %d, readySize: %d, status: %d", estimatedRequiredSize, readySize, status)
		if status == CacheStatusCompleted || readySize >= estimatedRequiredSize {
			break
		}

		// 检查是否超时
		if time.Since(startWait) > maxWaitTime {
			return fmt.Errorf("等待缓存超时，目标位置的数据尚未下载完成")
		}

		// 等待100ms后重试
		time.Sleep(100 * time.Millisecond)
	}

	// 发送跳转命令
	if err := p.sendCommand(fmt.Sprintf("JUMP %fs", position)); err != nil {
		return fmt.Errorf("跳转命令执行失败: %v", err)
	}

	// 等待跳转完成的反馈
	success := make(chan bool)
	timeout := time.After(2 * time.Second)

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := p.stdout.Read(buf)
			if err != nil {
				success <- false
				return
			}
			output := string(buf[:n])
			// mpg123 在成功跳转后会输出包含 "@J" 的消息
			if strings.Contains(output, "@J") {
				success <- true
				return
			}
		}
	}()

	select {
	case result := <-success:
		if !result {
			return fmt.Errorf("跳转操作失败")
		}
	case <-timeout:
		return fmt.Errorf("跳转操作超时")
	}

	log.Printf("[SeekTo] 成功跳转到 %.2f 秒", position)
	return nil
}

// GetCachedFile 获取缓存的文件信息
func (c *AudioCache) GetCachedFile(url string) *CacheInfo {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if info, exists := c.files[url]; exists {
		info.LastAccess = time.Now()
		return info
	}
	return nil
}

// initPlayer 初始化播放器
func (p *AudioPlayer) initPlayer() error {
	if p.cmd != nil {
		return nil
	}

	cmd := exec.Command("mpg123", "-R")
	var err error

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("无法获取标准输入管道: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("无法获取标准输出管道: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("无法启动 mpg123: %v", err)
	}

	p.cmd = cmd
	p.stdin = stdin
	p.stdout = stdout

	return nil
}

// sendCommand 发送命令到 mpg123
func (p *AudioPlayer) sendCommand(cmd string) error {
	if p.stdin == nil {
		return fmt.Errorf("播放器未初始化")
	}
	_, err := fmt.Fprintf(p.stdin, "%s\n", cmd)
	return err
}

// Pause 暂停播放
func (p *AudioPlayer) Pause() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isPlaying {
		return fmt.Errorf("没有正在播放的音频")
	}

	if err := p.sendCommand("PAUSE"); err != nil {
		return fmt.Errorf("暂停失败: %v", err)
	}

	return nil
}

// Resume 继续播放
func (p *AudioPlayer) Resume() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isPlaying {
		return fmt.Errorf("没有正在播放的音频")
	}

	if err := p.sendCommand("PAUSE"); err != nil { // mpg123 的 PAUSE 命令是切换暂停/继续状态
		return fmt.Errorf("继续播放失败: %v", err)
	}

	return nil
}

// Stop 停止播放
func (p *AudioPlayer) Stop() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.cmd != nil {
		p.sendCommand("STOP")
		p.sendCommand("QUIT")
		p.cmd.Process.Kill()
		p.stdin = nil
		p.stdout = nil
		p.cmd = nil
	}
	p.isPlaying = false
}
