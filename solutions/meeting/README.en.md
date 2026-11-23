# Meeting interpretation

[Solutions index](../README.en.md) · [中文](README.md)

## Current scope

The public community edition provides captions and translated audio from one microphone or WAV stream and supports Chinese–English turn-taking on one device. The official product also offers multi-party audio RTC meetings with translated audio routed by listener language, subject to environment and account access. Multi-party meetings are disabled in the community edition. [Community release scope](../../docs/open-source-plan.md) · [Official product links](../../README.en.md#official-web-app-and-product-pages)

## Verify the single-stream workflow

1. Start the community Web app and open Quick Start. Use the [lecture WAV](../classroom/sample-zh.wav) as the presenter's audio track.
2. Run the default simulation to verify captions, saving, and replay. To verify real translation, configure a community model provider or run the classroom example directly against interpretation v3.
3. Record the language direction, whether captions and translated audio are correct, and whether the final segment arrives after stopping.

[Deployment guide](../../docs/deployment.md) · [Runnable single-stream v3 example](../classroom/README.en.md) · [Community verification scope](../../docs/validation.md)

## Multi-party reference architecture

Ingest separate speaker tracks in the meeting system. Send the required tracks from your server to the interpretation API, then distribute captions and translated audio by listener language. Preserve the original audio and keep synthesized tracks out of the translation input. Verify join and leave, listener language switching, long sessions, and concurrency in your own integration. [Interpretation v3 reference](https://doc.vuilabs.cn/api-reference/realtime-interpretation/)
