# 听课 Go Quickstart 真实 v3 验收

[中文听课指南](README.md) · [English](#live-v3-validation)

2026-11-23 在 macOS arm64 上使用 Go 1.26.10，按听课指南的命令将仓库内 13.81 秒的[合成中文讲座](sample-zh.wav)发送给国内同传 v3，目标语言为英文。

| 模式 | 验收结果 | 客户端总耗时 |
|---|---|---:|
| `-text-only` | 收到完整的英文译文，正常结束，退出码为 `0` | 14.61 秒 |
| `-voice michael` | 收到相同的英文译文，退出码为 `0`。译音可解码，为 24 kHz 单声道 PCM16 WAV，时长 7.76 秒 | 15.03 秒 |

客户端总耗时包含连接、按播放速度发送音频和收尾。英文输出保留了原讲稿关于声音、空气传播和麦克风将振动转换为电信号的三项内容。[实际英文译文](translation-en.txt) · [实际英文译音](translation-en.wav) · [v3 协议](https://doc.vuilabs.cn/api-reference/realtime-interpretation/)

## Live v3 validation

[English classroom guide](README.en.md) · [中文](#听课-go-quickstart-真实-v3-验收)

On 2026-11-23, the commands in the classroom guide sent the repository's [13.81-second synthetic Mandarin lecture](sample-zh.wav) to mainland China interpretation v3 for English output. The client ran on macOS arm64 with Go 1.26.10.

| Mode | Result | Client wall time |
|---|---|---:|
| `-text-only` | Completed with exit code `0` and returned the English translation | 14.61 s |
| `-voice michael` | Exited with code `0`, returned identical text, and produced a decodable 7.76-second, 24 kHz mono PCM16 WAV | 15.03 s |

Client wall time includes connection setup, streaming at playback speed, and session finalization. The English output covers the three points in the source script about sound, how it travels through air, and how microphones turn vibrations into electrical signals. [Captured English text](translation-en.txt) · [Captured English speech](translation-en.wav) · [v3 protocol](https://doc.vuilabs.cn/api-reference/realtime-interpretation/)
