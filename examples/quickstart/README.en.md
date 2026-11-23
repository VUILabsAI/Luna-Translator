# Go API Quickstart — Luna interpretation v3

[中文](README.md) · English

This CLI calls the [VUI Labs mainland China interpretation v3 API](https://doc.vuilabs.cn/api-reference/realtime-interpretation/) directly. It streams a WAV file in real time, prints translation segments, and optionally saves translated speech in a preset voice. It does not require the community backend or enable voice cloning.

| Route | Endpoint and credential |
|---|---|
| This example: interpretation v3 | `wss://api.vuilabs.cn/v1/interpret/stream`; the handshake uses `X-API-Key` from `VUILABS_API_KEY`. |
| Existing community provider adapter | Uses `LUNA_WS_URL` and `Authorization: Bearer`; see the [community protocol](../../docs/providers.md). This example does not change that adapter. Do not mix credentials or endpoints. |

## Run it

Install Go 1.26.10 or newer. From the repository root, convert a Chinese recording you are authorized to send to a standard RIFF WAV with mono, 16 kHz PCM16. The example accepts up to 14 minutes; start with a few seconds. Live calls send the recording to the v3 service and require a mainland China key with interpretation access and available quota. Set `VUILABS_API_KEY` through a local secret manager or protected environment variable. Do not put it in command-line arguments or commit it.

```sh
mkdir -p output
ffmpeg -i /path/to/original.m4a -ac 1 -ar 16000 -c:a pcm_s16le output/input.wav
go -C examples/quickstart run . -input ../../output/input.wav -source zh -target en -text-only
```

For speech output, confirm a v3 preset voice ID using the [voice catalog](https://doc.vuilabs.cn/api-reference/realtime-interpretation/#音色查询). The command below uses `michael` as an example. Choose a file that does not already exist.

```sh
go -C examples/quickstart run . -input ../../output/input.wav -source zh -target en -voice michael -output ../../output/translated.wav
```

Translations go to stdout. The `request_id` and diagnostics go to stderr. Keep the ID to [check usage after an interruption](https://doc.vuilabs.cn/api-reference/realtime-interpretation/#用量额度与中断). Authentication, quota, connection, and synthesis failures return a nonzero exit status. Failed synthesis does not save a partial WAV, and the CLI does not automatically resend audio.

## Verify without a key

```sh
go -C examples/quickstart test -race ./...
```

The tests use a local mock WebSocket to check the protocol and error handling. They do not call the live API or consume quota. Go may download dependencies on the first run.
