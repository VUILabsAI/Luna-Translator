# 模型服务与通信协议

本项目只公开应用及调用端。使用模型需要您自己的服务账号、API 权限及额度。

## Luna 公共 API

文档：https://simultrans.lunalabs.cn/api-docs

后端用 `Authorization: Bearer <your-key>` 连接 `LUNA_WS_URL`。服务端代码位于 `backend/providers/lunapublic/`，凭据不下发浏览器。

连接后发送 `action=start`，设置 `source_lang`、`target_lang` 和 `enable_tts=true`。社区版默认使用 `tts_backend=lunasi-tts-sc`、`voice_mode=preset`、`voice=avery`。收到 `started` 后，以实时节奏发送 16 kHz 单声道 PCM16 小端二进制块，每包最多 1 秒。一次连接只发送一次 `start`。两段录音之间可以继续送音，所有输入发完再发送一次 `action=finish`，继续接收到 `completed`。

### 热词与 VAD

可在本地 `.env` 设置以下选项，再重新运行启动脚本：

```dotenv
LUNA_ENABLE_VAD=false
LUNA_HOTWORDS='[{"source":"西湖","target":"West Lake"}]'
```

- `LUNA_HOTWORDS` 是结构化词对，每项只含 `source` 和 `target`。最多 20 对，每个词去除首尾空白后为 1–50 字符，同一个源词只出现一次。省略或设置 `[]` 使用标准翻译。词对作为部署配置，会附加到每条 Luna 上游调用。多目标语言部署宜使用各语种通用的固定译法。生成的完整提示词还受上游字符数和 token 限制。
- Bearer 接口拒绝 `system_prompt`，包括空字符串和 `null`。社区版发送结构化词对，旧的自由提示词无法设置翻译行为。
- `LUNA_ENABLE_VAD` 默认关闭，设为 `true` 开启推理侧 VAD。停顿后的输出时机由上游决定。连续 60 秒没有有效音频时，上游会自动收尾，WebSocket 心跳不算音频。发送的静音 PCM 仍按输入音频计量。详情见[官方 VAD 参数和用量说明](https://simultrans.lunalabs.cn/api-docs#parameters)。

`translation` 是增量文本，每条 `speech` / `speech_chunk` JSON 后是对应的 float32 小端二进制音频。适配器校验音频包长度、采样率和非有限样本，并转换为社区 Web 的 PCM16 事件。语音段编号独立于文本编号。`translation_complete` 后仍可能收到译音。如果 `completed` 标记 `partial`，社区版报告不完整输出。上游发送 `usage` 后结束，可通过[上游用量收据](https://simultrans.lunalabs.cn/api-docs#receipts)按 `session_id` 查询分项 token。

### 双段协议检查

`go -C backend test ./providers/lunapublic -run TestBearerTwoSegments -v -count=1` 使用本地模拟服务。在 VAD 开和关两组中分别通过同一条连接发送两段输入，校验两段译文、合成语音帧、独立语音段编号及模拟用量。真实复测时建议使用两句易区分且有权提交的录音，中间插入约 2 秒静音，两段发完才 `finish`。记录第二段的文本和对应 PCM、`usage`、`completed`，按同一 `session_id` 核对上游收据。

## 百炼

配置 `DASHSCOPE_API_KEY`、`DASHSCOPE_WORKSPACE_ID`、`DASHSCOPE_REGION`。原文识别通过 `qwen3-asr-flash-realtime`；可以使用 `DASHSCOPE_ASR_URL` 覆盖工作空间派生的 WebSocket 地址。Luna 原文与译文由两条独立模型链路生成，可能在措辞和时间边界上不完全一致。

向量评分使用 `text-embedding-v4` 的 OpenAI 兼容 `/embeddings` 接口，由 `DASHSCOPE_EMBEDDING_URL` 指定完整地址。请求包含本次原文与译文两段文本。当前应用只返回**整段近似配对与余弦相似度**；这不是逐句对齐准确率保证，也不包含内部 Embedding 对齐模型。

若需使用附带的百炼 LiveTranslate 适配器，可自行设置 `ENABLE_BAILIAN_TRANSLATE=true`；此替代翻译模型不作为首版默认链路，其真实权限与表现须自行验证。

## 浏览器到社区后端

1. `POST /api/community/login` 换取工作区 access token。
2. 带 Bearer 调用 `POST /api/consumer/ws/session`，指定 `provider`，获得 `sessionId` 与一次性 `ticket`。
3. 连接同源 `/ws/translate`，首个 JSON 为 `type=start`，包含上述票据、provider、sourceLanguage、targetLanguages。
4. 发送 16 kHz 单声道 PCM16 二进制。结束发送 `type=stop`，继续接收字幕、音频和 `state=finished`。

票据与 provider/sessionId 绑定，30 秒有效。输入语言为 zh/en；单次最多 3 个不重复目标语言；部署最多同时接收 4 条会话连接，每条最多 30 分钟，每方向音频内存上限 64 MiB。上游提供方的并发/额度可能更低。

模型未配置、鉴权失败、连接中断或超时时，应用返回错误；不会用模拟结果替代真实会话。`DEMO_MODE=true` 时只注册模拟服务，固定示例与提示音始终明确标注。

社区 v0.1 的单人字幕按整段、双向字幕按每轮发言展示，避免把独立 ASR 和译音的不同分句误配。Embedding 适配接口可独立调用与扩展，不能推断出已验证的逐句边界。
