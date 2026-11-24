<p align="center">
  <img src="web/public/welcome/assets/luna-logo.png" alt="Luna Translator" width="88" />
</p>

<h1 align="center">Luna Translator</h1>

<p align="center">
  VUI Labs 的实时语音同传开源应用与 API 解决方案，提供字幕、译音和可自托管的 Web 社区版。
</p>

<p align="center">
  中文 · <a href="README.en.md">English</a>
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache_2.0-blue.svg" alt="License: Apache-2.0" /></a>
  <a href="https://github.com/VUILabsAI/Luna-Translator/actions/workflows/ci.yml"><img src="https://github.com/VUILabsAI/Luna-Translator/actions/workflows/ci.yml/badge.svg?branch=main" alt="Community checks" /></a>
  <a href="https://github.com/VUILabsAI/Luna-Translator/releases/latest"><img src="https://img.shields.io/github/v/release/VUILabsAI/Luna-Translator" alt="最新社区版发行版本" /></a>
  <a href="#安装与快速开始"><img src="https://img.shields.io/badge/Docker%20Compose-self--hosted-2496ED?logo=docker" alt="通过 Docker Compose 自托管" /></a>
  <br />
  <a href="https://github.com/VUILabsAI/Luna-Translator"><img src="https://img.shields.io/github/stars/VUILabsAI/Luna-Translator?style=flat" alt="GitHub Stars 计数" /></a>
  <img src="https://img.shields.io/github/languages/top/VUILabsAI/Luna-Translator" alt="仓库主要语言" />
</p>

| 使用 Luna | 接入 Luna |
|---|---|
| [![在线体验](https://img.shields.io/badge/%E5%9C%A8%E7%BA%BF%E4%BD%93%E9%AA%8C-0969DA?style=for-the-badge)](https://simultrans.lunalabs.cn/) | [![Go API 快速开始](https://img.shields.io/badge/Go%20API%20%E5%BF%AB%E9%80%9F%E5%BC%80%E5%A7%8B-007D9C?style=for-the-badge&logo=go&logoColor=white)](examples/quickstart/README.md) |
| [自托管 Web 社区版](#安装与快速开始) · [官方网页版与产品介绍](#官方网页版与产品介绍) | [按场景接入](solutions/README.md) · [同传 v3 API 文档](https://doc.vuilabs.cn/api-reference/realtime-interpretation/) |
| 社区版可在本机运行模拟体验，无需模型 API Key。 | Go 示例直连国内同传 v3，使用 `X-API-Key`。社区版通过 [Bearer 模型适配器](docs/providers.md#luna-公共-api)接入自己的模型服务。 |

## 边说，边译，边听

听一场外语分享，向不同语言的听众介绍想法，或与面前的人轮流交谈——Luna Translator 将输入语音转换为原文字幕、目标语言译文与译音，并保留会话，方便之后回听和分享。

这个仓库提供可独立部署的 **Web 前端与 Go 后端**。你可以在自己的电脑或服务器上运行，修改界面、扩展业务流程，或接入自己的模型服务。应用通过外部 API 完成实际识别与翻译，本机部署无需 GPU。

> 当前社区版优先支持 Web：单人同传、中英双向对话、历史、回放与分享。默认启动的是有明确标识的模拟体验；实际识别和翻译需要配置自己的模型服务。

### 社区版演示

[![Web 社区版模拟会话记录：中英示例字幕与音频回放控件](docs/assets/community-web-demo.jpg)](docs/assets/community-web-demo.mp4)

[观看 Web 社区版录屏（MP4）](docs/assets/community-web-demo.mp4)：用 WAV 输入启动模拟同传，查看字幕和保存的会话。示例字幕由模拟模式生成，不是对输入内容的实际翻译。

## 最新动态

- **main 分支** · 已核对 iOS 1.0.11（130）与 Android 86。历史详情可重新检查暂缺的音轨，保留已可播放的音频。此前的终态记录保护也包含在 `main` 中。社区版 Web 沿用 iOS 1.0.11（128）的界面样式，`v0.1.1` 下载包对应此前的发行提交。
- **2026-11-23 · [v0.1.1](https://github.com/VUILabsAI/Luna-Translator/releases/tag/v0.1.1)** · 新增四个双语场景指南和 Go Quickstart。听课示例附带真实中译英的译文和译音。公开发行包已通过干净环境的 Docker Compose 启动、WAV 输入和浏览器回放验收。[验收记录](docs/validation.md#v011)
- **2026-10-30 · [v0.1.0](https://github.com/VUILabsAI/Luna-Translator/releases/tag/v0.1.0)**：发布首个 Web + Go 后端社区版，提供 Docker Compose 启动、Luna / 百炼适配、历史与音频归档、可撤销分享，以及独立的公开 CI。

## 从实时交流，到回听分享

| 场景 | 如何使用 |
|---|---|
| 单人同传 | 选择输入和目标语言，查看原译文并播放译音；一次最多选择 3 种目标语言。 |
| 面对面对话 | 中英双方轮流发言，切换当前发言方，上下两侧分别查看对应语言字幕。 |
| 会后回听 | 打开历史，查看已保存文字，播放或下载原音与译音 WAV。 |
| 分享记录 | 为历史生成限时快照链接，接收者无需工作区密码即可查看；可随时撤销。 |
| 开发与联调 | 无模型密钥时走模拟流程，也可用 PCM16 WAV 按实时速度复现同传链路。 |

### 核心能力

- **实时字幕与译音**：支持中文 / 英文输入；目标语言由 Luna 公共 API 提供，最多同时输出 3 种。
- **单人和双向交互**：快速开始或中英轮流对话，保留双方的文字与音轨。
- **完整会话记录**：本地数据卷持久保存历史与音频，支持浏览、回放、下载和删除。
- **可撤销分享**：生成最长 30 天的分享快照；撤销链接或删除原记录后，链接失效。
- **可扩展服务接入**：通过应用层适配器调用 Luna 同传、百炼 ASR 与 Embedding。
- **独立部署**：两个容器、一个数据卷、自动生成工作区密码，无需外部数据库或短信平台。

## 参考架构

![Luna Translator 社区版架构：浏览器、Go 后端、外部模型 API 与持久化数据卷](docs/assets/architecture.svg)

模型凭据保留在后端。浏览器负责采集和播放，后端负责服务编排与数据保存。完整接口和数据生命周期见 [架构说明](docs/architecture.md) 与 [服务协议](docs/providers.md)。

## 安装与快速开始

### 1. 下载项目

安装 [Docker Engine / Docker Desktop 与 Compose](https://docs.docker.com/compose/install/)，然后克隆仓库：

```sh
git clone https://github.com/VUILabsAI/Luna-Translator.git
cd Luna-Translator
```

也可以从 [Releases](https://github.com/VUILabsAI/Luna-Translator/releases) 下载 ZIP 或 tar.gz 并解压；同页提供 SHA-256 校验文件。

### 2. 一条命令启动

macOS / Linux：

```sh
sh ./start.sh
```

Windows PowerShell：

```powershell
./start.ps1
```

脚本会构建 Web 与后端镜像、等待健康检查，并显示本部署的工作区密码。打开 <http://localhost:8088> 登录，即可体验。

首次构建需要访问公开镜像及 npm / Go 依赖源；Docker 部署无需本机安装 Node.js 或 Go。需要自定义端口、远程 HTTPS、备份或升级时，参见 [部署指南](docs/deployment.md)。

> **默认模拟模式**只显示固定示例字幕和提示音，不会识别或翻译实际语音。模拟音频进入本机后端；接入真实服务后，音频才会发送给所配置的提供方。

### 3. 试用一次会话

选择输入和输出语言，点击「开始同传」，完成音频授权后开始说话。也可先选择无压缩 PCM16 WAV，按实时速度发送测试音频。会话完成后，可以查看历史、回放音频或生成分享链接；面对面交流时切换「对话」模式。

## 接入模型服务

各项能力可以按需配置：

| 能力 | 推荐服务 | 接入准备 | 在社区版中的用途 |
|---|---|---|---|
| 同传 | [Luna 公共 API](https://simultrans.lunalabs.cn/api-docs) | Luna API Key | 实时译文与合成语音；固定音色，默认不开启声音克隆。 |
| 原文识别 | [百炼 `qwen3-asr-flash-realtime`](https://help.aliyun.com/zh/model-studio/real-time-speech-recognition-user-guide) | API Key、工作空间与地域 | 为 Luna 链路补充原文字幕。 |
| 向量评分 | [百炼 `text-embedding-v4`](https://help.aliyun.com/zh/model-studio/text-embedding-synchronous-api) | API Key 与 Embedding 完整接口地址 | 通过后端接口计算整段原译文相似度，供开发者调用与扩展。 |

先复制配置示例（macOS / Linux 使用 `cp .env.example .env`；PowerShell 使用 `Copy-Item .env.example .env`），再编辑 `.env`：

```dotenv
# 开启真实同传；只配置 Luna 时即可获得译文和译音。
DEMO_MODE=false
LUNA_API_KEY=your-own-luna-key

# 可选：补充原文识别。
DASHSCOPE_API_KEY=your-own-model-studio-key
DASHSCOPE_WORKSPACE_ID=your-workspace-id
DASHSCOPE_REGION=cn-beijing

# 可选：启用向量评分接口，地址以自己的工作空间配置为准。
DASHSCOPE_EMBEDDING_URL=https://your-workspace-id.cn-beijing.maas.aliyuncs.com/compatible-mode/v1/embeddings
```

重新运行启动脚本以应用配置。未配置百炼 ASR 时，Luna 译文和译音仍可使用，原文字幕不可用。服务开通、费用、地域和并发限制由各提供方决定；多个目标语言会分别建立上游调用并占用额度。

社区版按整段或每轮发言展示原译文。Embedding 提供向量评分，不能代替已验证的逐句对齐模型。更多参数和适配方式见 [模型配置与协议](docs/providers.md)。自己的 `.env` 和密钥应留在本机，不提交到仓库或公开 Issue。

## 本地开发

需要 **Node.js 22.12+、Go 1.26.10+**。在仓库根目录安装依赖：

```sh
npm ci
```

macOS / Linux 在一个终端启动模拟后端：

```sh
DATA_DIR="$PWD/data" DEMO_MODE=true go -C backend run .
```

在另一个终端启动前端：

```sh
npm run dev --workspace web
```

打开 <http://localhost:5173>。首次生成的工作区密码位于 `data/admin-password`；前端通过 Vite 代理连接本机 `8080` 后端。

构建与检查：

```sh
npm run build
npm test
go -C backend test -race ./...
python3 scripts/check-public-content.py
```

| 目录 | 内容 |
|---|---|
| `web/` | Web 页面、音频交互、历史与分享界面。 |
| `packages/core/`、`packages/ui-shared/` | 共享文本状态、协议与浏览器音频逻辑。 |
| `backend/community/` | 独立工作区 API、会话编排与持久化。 |
| `backend/providers/` | 外部模型服务客户端适配器。 |
| [`solutions/`](solutions/README.md) | 会议、听课、直播和通话的场景指南与接入范围。 |
| [`examples/quickstart/`](examples/quickstart/README.md) | 可运行的国内同传 v3 Go 示例。 |
| `docs/` | 架构、部署、服务接入、验收与路线图。 |

## 版本范围与路线图

社区版面向**单工作区部署**。使用同一部署密码的人共享历史与模型额度。不同用户需要独立部署，远程浏览器麦克风需要有效 HTTPS。

已完成真实合成语音的中译英 / 英译中、浏览器历史与分享、容器启动和公开 CI 验证。真人麦克风、耳机 / 回声、Safari / Edge、Windows 实机及长会话负载尚未完成验收，具体记录见 [验证说明](docs/validation.md)。

后续优先完善 Web 的兼容性、可访问性、模型适配和对齐策略，再扩展多人会议 / RTC，整理 Android 与 iOS 源码。完整安排见 [开源计划](docs/open-source-plan.md)。

## 官方网页版与产品介绍

| 地区 | 产品介绍（Welcome） | 官方网页版 |
|---|---|---|
| 中国大陆 | [大陆产品页](https://simultrans.lunalabs.cn/welcome/) | [打开大陆网页版](https://simultrans.lunalabs.cn/) |
| 国际通用入口 | [国际产品页（英文）](https://simultrans.vuilabs.ai/welcome/?lang=en) | [打开国际网页版](https://simultrans.vuilabs.ai/) |
| 美国（北美） | [北美产品页（英文）](https://simultrans.vuilabs.ai/us/welcome/?lang=en) | [打开北美网页版](https://simultrans.vuilabs.ai/us/) |

官方网页版与独立部署的社区工作区使用各自的账号和数据。国际和北美网页版需要先登录。官方 API 的能力和协议另见[国内同传 API 文档](https://doc.vuilabs.cn/api-reference/realtime-interpretation/)。

### 官方移动端预览

<p>
  <img src="docs/assets/official-ios-preview.jpg" alt="官方 iOS App：快速同传入口（模拟器截图）" width="205" />
  <img src="docs/assets/official-android-preview.jpg" alt="官方 Android App：快速同传入口（模拟器截图）" width="205" />
</p>

上图为官方 App 的 iOS 和 Android 模拟器截图，不代表移动端源码已发布。中国大陆 iOS 用户可前往 [App Store](https://apps.apple.com/cn/app/id6806330215)，社区版进展见[开源路线图](docs/open-source-plan.md#移动端发布准备)。

## FAQ 与排障

- **没有 API Key 能运行吗？** 可以。默认模拟模式无需模型凭据，可体验字幕、历史和分享流程；真实翻译见[模型接入](#接入模型服务)。
- **启动失败或端口被占用？** 用 `docker compose logs` 查日志，必要时在 `.env` 中修改 `PORT`。更多问题见[部署排障](docs/deployment.md#常见问题)。
- **无法使用麦克风？** 检查浏览器权限；远程访问需要 HTTPS。也可选择无压缩 PCM16 WAV 文件测试。

## 交流与贡献

欢迎通过 [Issues](https://github.com/VUILabsAI/Luna-Translator/issues) 反馈问题、讨论场景和提出功能建议，也欢迎提交 Pull Request。当前适合参与的方向包括浏览器兼容、无障碍、模型适配、文档与对齐策略。

- [贡献指南](CONTRIBUTING.md)：开发检查与提交说明。
- [安全报告](SECURITY.md)：私有凭据、录音及安全问题的报告方式。
- [第三方声明](THIRD_PARTY_NOTICES.md)：依赖、字体与素材许可。

## 许可证

应用代码采用 [Apache License 2.0](LICENSE)，第三方组件遵循各自许可。本仓库不分发 ASR、同传、Embedding 对齐等模型的权重、训练代码、训练数据或内部模型推理实现；模型服务的使用条件由提供方规定。
