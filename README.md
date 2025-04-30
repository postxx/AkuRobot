# AkuRobot Web Interface

一个功能丰富的 Web 音乐播放器和系统管理界面，支持多种音乐播放模式和系统服务管理。

## 主要功能

### 音乐播放功能

1. **网易云音乐播放器** (NetMusic.html)
   - 支持网易云音乐歌单导入和播放
   - 实时歌单加载和动态分页
   - 完整的播放控制（播放/暂停/上一首/下一首）
   - 音量控制和进度条拖拽
   - VIP 歌曲标识
   - 优雅的加载动画和状态提示

2. **本地音乐播放器** (music_local.html)
   - 支持播放本地音乐文件
   - 支持 MP3 和 WAV 格式
   - 基础播放控制功能

3. **流媒体播放器** (music_url.html)
   - 支持在线音乐流播放
   - 实时音频流控制

### 系统管理功能

1. **服务管理** (service.html)
   - 第三方系统服务的状态监控
   - 服务启动/停止控制
   - 实时服务输出日志查看
   - 服务状态实时更新

2. **系统监控** (system.html)
   - 系统信息和硬件状态查看
   - 系统重启功能
   - 硬件资源监控

## 技术特点

1. **音频播放功能**
   - 支持多种音频格式
   - 精确的音量控制
   - 播放进度控制
   - 实时状态同步

2. **用户界面**
   - 响应式设计
   - 现代化 UI 界面
   - 流畅的动画效果
   - 直观的操作反馈

3. **后端功能**
   - RESTful API 设计
   - 实时事件流处理
   - 服务状态监控
   - 文件系统管理

## API 接口

### 音乐相关接口
- `/api/music/list` - 获取音乐列表
- `/api/music/stream` - 流媒体播放
- `/api/music/stop` - 停止播放
- `/api/music/pause` - 暂停播放
- `/api/music/resume` - 继续播放
- `/api/music/seek` - 播放进度控制
- `/api/volume/get` - 获取音量
- `/api/volume/set` - 设置音量

### 网易云音乐接口
- `/api/playlist/detail` - 获取歌单详情
- `/api/playlist/play` - 播放歌单歌曲

### 系统管理接口
- `/api/service/start` - 启动服务
- `/api/service/stop` - 停止服务
- `/api/service/status` - 获取服务状态
- `/api/service/output` - 获取服务输出
- `/api/system/reboot` - 系统重启

## 安装和使用

1. 克隆项目
```bash
git clone https://github.com/yourusername/AkuRobot.git
```

2. 安装依赖
```bash
# 确保系统已安装必要的音频库
sudo apt-get install mpv
```

3. 运行服务
```bash
go run main.go
```

4. 访问界面
打开浏览器访问 `http://localhost:8080`

## 配置说明

主要配置项在 `config` 包中定义：
- 默认目录设置
- 服务配置
- 音频播放器配置

## 注意事项

1. 网易云音乐功能需要：
   - 有效的网络连接
   - 合法的音乐版权
   - 非 VIP 歌曲限制

2. 系统管理功能需要：
   - 适当的系统权限
   - 相关服务的安装和配置

## 贡献指南

欢迎提交 Issue 和 Pull Request 来帮助改进项目。

## 许可证

[MIT License](LICENSE)

## 作者

[Your Name]

## 更新日志

### v1.0.0
- 初始版本发布
- 基础音乐播放功能
- 系统管理功能

## 鸣谢

- MPV 播放器
- 网易云音乐 API
