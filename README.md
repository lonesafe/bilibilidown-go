# BilibiliDown Go + Vue3

> B站视频下载器 — 从 Java Swing 版反编译后重构为 **Go 后端 + Vue3 前端** 架构

原版作者: [nICEnnnnnnnLee/BilibiliDown](https://github.com/nICEnnnnnnnLee/BilibiliDown)

---

## ✨ 功能特性

- 🎬 **视频解析** — 支持 BV/AV 号及各种 B站链接，自动获取视频信息、分P列表、多清晰度
- ⬇️ **多线程下载** — DASH 音视频流下载，支持并发、进度/速度实时上报
- 🔀 **FFmpeg 合并** — 自动合并视频流与音频流为 MP4
- 📱 **扫码登录** — 生成二维码，手机 B站 App 扫码自动获取 Cookie
- 🍪 **Cookie 登录** — 支持手动粘贴 Cookie 登录
- 🔐 **WBI 签名** — 完整移植原版 B站 WBI URL 签名算法
- 🔢 **BV/AV 互转** — 原版加密算法完整移植
- ⚡ **实时推送** — WebSocket 推送下载进度（500ms 间隔）
- ⚙️ **配置管理** — 保存路径、FFmpeg、并发数、编码优先级等

---

## 🛠 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.21, gorilla/websocket, tidwall/gjson |
| 前端 | Vue 3, Vite, Element Plus, Vue Router, Axios, qrcode |
| 外部依赖 | FFmpeg（音视频合并） |

---

## 📁 项目结构

```
bilibilidown-go/
├── backend/                          # Go 后端
│   ├── cmd/server/main.go            # 入口
│   ├── go.mod / go.sum
│   └── internal/
│       ├── api/
│       │   ├── bilibili.go           # B站API客户端（视频详情/播放链接/字幕/弹幕）
│       │   └── login.go              # 扫码登录（QR生成+轮询+Cookie获取）
│       ├── parser/parser.go          # URL解析器（BV/AV/SS/EP/ML/b23）
│       ├── downloader/downloader.go  # 下载管理器（并发+ffmpeg合并+进度）
│       ├── model/video.go            # 数据模型
│       ├── config/config.go          # 配置管理
│       ├── server/server.go          # HTTP服务器 + WebSocket实时推送
│       └── util/
│           ├── wbi.go                # WBI URL签名
│           ├── convert.go            # BV/AV ID互转
│           ├── cookie.go             # Cookie管理
│           ├── headers.go            # HTTP请求头
│           ├── http.go               # HTTP客户端
│           └── file.go               # 文件工具
└── frontend/                         # Vue3 前端
    ├── index.html
    ├── package.json
    ├── vite.config.js
    └── src/
        ├── main.js / App.vue / router.js / api.js
        └── views/
            ├── Home.vue              # 视频解析
            ├── Downloads.vue         # 下载任务
            ├── Settings.vue          # 设置
            └── Login.vue             # 登录（扫码+Cookie）
```

---

## 🚀 快速开始

### 环境要求

- Go 1.21+
- Node.js 16+
- FFmpeg（系统 PATH 中可用，或在设置中指定路径）

### 1. 构建前端

```bash
cd frontend
npm install
npm run build
```

### 2. 编译并运行后端

```bash
cd backend
go build -o bilibilidown-server ./cmd/server/
./bilibilidown-server
```

### 3. 访问

打开浏览器访问 **http://localhost:8787**

> 后端服务默认监听 `8787` 端口，同时托管前端静态文件。

---

## 📖 使用说明

### 登录

前往「登录」页面，两种方式任选其一：

1. **扫码登录（推荐）** — 自动生成二维码，用 B站 App 扫码并在手机上确认登录，后端自动捕获 Cookie。
2. **Cookie 登录** — 从浏览器复制 B站 Cookie 粘贴到输入框。

### 下载视频

1. 在「视频解析」页面输入 BV/AV 号或视频链接
2. 选择清晰度，点击「解析」
3. 在分P列表中选择清晰度，点击「下载」
4. 在「下载任务」页面查看实时进度

### 设置

在「设置」页面可配置：
- 下载保存路径
- FFmpeg 路径
- 并发下载数
- 下载失败重试次数
- 下载格式（MP4/FLV）
- 视频编码优先级（HEVC/AVC/AV1）
- 音频质量优先级
- HTTP 服务端口

---

## 🔌 API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/parse` | 解析视频（body: `{url, qn, fetchLinks}`） |
| POST | `/api/download` | 创建下载任务 |
| GET | `/api/tasks` | 获取所有下载任务 |
| POST | `/api/tasks/stop` | 停止下载任务 |
| GET | `/api/login/status` | 登录状态 |
| POST | `/api/login/cookie` | 设置 Cookie |
| GET | `/api/login/qrcode` | 生成登录二维码 |
| POST | `/api/login/qrcode/poll` | 轮询扫码状态 |
| GET | `/api/user` | 获取用户信息 |
| GET/POST | `/api/config` | 获取/更新配置 |
| WS | `/api/ws` | WebSocket 实时推送下载进度 |

---

## ⚠️ 声明

本项目仅供学习交流使用，请勿用于商业用途。下载的视频版权归原作者及 Bilibili 所有，请遵守相关法律法规。
