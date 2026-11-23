# Luna 同传解决方案

[项目首页](../README.md) · [English](README.en.md)

按业务场景选择公开示例或参考架构。各场景页列明可用版本和验证范围。

| 场景 | 当前可用范围 |
|---|---|
| [听课与讲座](classroom/README.md) | **可运行示例。** 用随附的中文讲课音频调用同传 v3，查看英文译文并保存译音。 |
| [会议同传](meeting/README.md) | **社区版单路流程。** 主讲人可用 Web 获取字幕和译音。多人 RTC 属于官方产品能力，本页提供接入参考。 |
| [直播同传](livestream/README.md) | **Mac OBS 试点。** 公开仓库提供单路翻译验证路径，OBS 和实际平台推流的验收单独记录。 |
| [通话同传](calls/README.md) | **参考架构。** 公开社区版可验证面对面轮流对话，第三方音轨回注与电话／SIP 接入需要独立实现。 |

## 选择接入方式

- **运行开源应用。** 用 Docker Compose 启动 Web + Go 社区版。默认模拟模式可免 Key 跑通流程，真实翻译使用部署者配置的模型服务。Luna 社区适配器使用 `Authorization: Bearer`。[部署指南](../docs/deployment.md) · [社区版协议](../docs/providers.md)
- **将同传接入自己的系统。** 用 Go Quickstart 直连国内同传 v3，握手使用 `X-API-Key`。Web 集成应由自己的服务端持有 Key。[Go Quickstart](../examples/quickstart/README.md) · [同传 v3 文档](https://doc.vuilabs.cn/api-reference/realtime-interpretation/)
