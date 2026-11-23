# Go API Quickstart — Luna 同传 v3

中文 · [English](README.en.md)

这个命令行示例直接接入 [VUI Labs 国内同传 v3](https://doc.vuilabs.cn/api-reference/realtime-interpretation/)，以实时速度发送 WAV 音频，逐段打印译文，并可保存预置音色的译音。无需启动社区版后端；示例不启用声音克隆。

| 链路 | 端点与凭据 |
|---|---|
| 本示例：同传 v3 | `wss://api.vuilabs.cn/v1/interpret/stream`；握手使用 `X-API-Key`，从 `VUILABS_API_KEY` 读取。 |
| 当前社区版的模型适配器 | 使用 `LUNA_WS_URL` 与 `Authorization: Bearer`，见[社区版协议](../../docs/providers.md)。不因本示例而改变，凭据和端点不要混用。 |

## 运行

安装 Go 1.26.10 或更新版。在仓库根目录准备一段有权提交的中文录音，并用 FFmpeg 转成普通 RIFF PCM16、16 kHz、单声道 WAV。此示例最多接受 14 分钟，建议首次用几秒钟的音频。真实调用会将录音发送到 v3 服务，需要已开通国内同传权限的 Key 和可用额度。通过本地密钥管理器或受保护的环境变量设置 `VUILABS_API_KEY`，不要把 Key 提交到仓库或放进命令参数。

```sh
mkdir -p output
ffmpeg -i /path/to/original.m4a -ac 1 -ar 16000 -c:a pcm_s16le output/input.wav
go -C examples/quickstart run . -input ../../output/input.wav -source zh -target en -text-only
```

需要译音时，从[音色查询](https://doc.vuilabs.cn/api-reference/realtime-interpretation/#音色查询)确认对应的 v3 预置音色 ID。以下用 `michael` 为例；输出文件必须不存在。

```sh
go -C examples/quickstart run . -input ../../output/input.wav -source zh -target en -voice michael -output ../../output/translated.wav
```

译文写入标准输出，`request_id` 和诊断信息写入标准错误。保留 `request_id`，用于中断后的[用量查询](https://doc.vuilabs.cn/api-reference/realtime-interpretation/#用量额度与中断)。鉴权、额度、网络和合成失败会以非零状态退出；合成失败不会保存部分 WAV。程序不会自动重放音频。

## 无 Key 验证

```sh
go -C examples/quickstart test -race ./...
```

测试用本地模拟 WebSocket 验证协议和错误处理，不连接真实 v3、不消耗额度。首次运行可能需要下载 Go 依赖。
