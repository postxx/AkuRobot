package config

// Server 配置
const (
	DefaultPort = "80"
	DefaultDir  = "static"
)

// 音频相关配置
const (
	MaxVolume = 63
)

// 小智AI服务配置
const (
	XiaozhiSoundPath = "/opt/aku/xiaozhi/XIAOZHI_AI_SOUND" // 小智AI声音服务路径
	XiaozhiMainPath  = "/opt/aku/xiaozhi/XIAOZHI_AI_MAIN"  // 小智AI主服务路径
	XiaozhiGuiPath   = "/opt/aku/xiaozhi/XIAOZHI_AI_GUI"   // 小智AI GUI服务路径
)

// 底包程序配置
const (
	ShowImgPath   = "/opt/aku/web/show_image"        // 显示图片程序路径
	ShowGifPath   = "/opt/aku/web/play_bmp_sequence" // 播放GIF动画程序路径
	ShowTextPath  = "/opt/aku/web/show_text"         // 显示文字程序路径
)
