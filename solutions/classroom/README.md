# 听课与讲座同传

[解决方案索引](../README.md) · [English](README.en.md)

用现有 Go Quickstart 将讲课录音按播放速度流式发送给国内同传 v3。终端逐段显示英文译文，也可以把译音保存为 WAV。该示例覆盖一条主讲人音轨。官方产品的一对多实时分享可供架构参考，学生端收听和课堂管理需要另行接入。

## 运行示例

安装 Go 1.26.10 或更新版本。在本机环境中配置 `VUILABS_API_KEY`，使用已开通国内同传 v3 权限且有可用额度的 Key。示例会把音频发送给 v3 服务。随附的 [中文讲座 WAV](sample-zh.wav) 由 eSpeak NG 合成，格式为 16 kHz 单声道 PCM16。[Go Quickstart](../../examples/quickstart/README.md)

在仓库根目录运行，先确认可以收到译文。

```sh
go -C examples/quickstart run . -input ../../solutions/classroom/sample-zh.wav -source zh -target en -text-only
```

要生成译音，在同一位置运行以下命令。`michael` 是示例预置音色，可按音色查询结果替换。请为每次运行选择一个尚不存在的输出文件。[音色查询](https://doc.vuilabs.cn/api-reference/realtime-interpretation/#音色查询)

```sh
mkdir -p output
go -C examples/quickstart run . -input ../../solutions/classroom/sample-zh.wav -source zh -target en -voice michael -output ../../output/classroom-zh-en.wav
```

成功后，终端会显示译文，标准错误会给出 `request_id`，第二条命令还会生成 `output/classroom-zh-en.wav`。样例讲的是声音如何在空气中传播，译文措辞随模型输出变化。[用量与中断查询](https://doc.vuilabs.cn/api-reference/realtime-interpretation/#用量额度与中断)

## 真实 v3 验收结果

同一段讲座音频通过了仅译文和预置音色两种模式。两次输出的英文一致，带译音模式生成了 7.76 秒的 WAV。[英文译文](translation-en.txt) · [英文译音](translation-en.wav) · [验收记录](validation.md)

## 无 Key 验证

```sh
go -C examples/quickstart test -race ./...
```

测试用本地模拟 WebSocket 核验协议与失败处理，不调用真实模型。完整课程接入时，由服务端持有 v3 Key、接收讲师音轨并分发结果。长课应按服务的会话时长与并发限制设计切换及收尾。[同传 v3 文档](https://doc.vuilabs.cn/api-reference/realtime-interpretation/)
