# Livestream interpretation

[Solutions index](../README.en.md) · [中文](README.md)

## Current scope

The public repository covers single-stream audio input, streaming translated text, and translated audio. A Mac product pilot uses a Mandarin microphone track and sends English, Japanese, and Korean captions to local OBS browser sources. Packaged builds, real OBS operation, and streaming to live platforms still require separate verification. This page provides the reference workflow. [Single-stream v3 example](../classroom/README.en.md)

## Connect one audio stream to a livestream

1. Use the included WAV to verify translated text in the Go Quickstart and generate translated audio if needed.
2. Take an authorized presenter track from a microphone, mixer, or production tool and convert it to the mono PCM16 required by v3.
3. Send translated text from your server to your caption overlay. Let OBS or your production tool compose and stream the video. For translated audio, use a separate track and handle mixing and synchronization there.

Before going live, verify source-to-caption sync, pause and restart, interrupted inputs, and session rollover during long streams. Record caption-source tests separately from actual broadcast tests. [Interpretation v3 audio and events](https://doc.vuilabs.cn/api-reference/realtime-interpretation/#发送音频与读取结果)
