# Call interpretation

[Solutions index](../README.en.md) · [中文](README.md)

## Current scope

The public community edition provides Chinese–English face-to-face turn-taking to verify two translation directions and playback on one device. The official product's multi-party meetings are a reference for in-app remote voice. This page outlines how to integrate third-party voice or video calls, translated audio injection, and phone or SIP media in your own system. [Community release scope](../../docs/open-source-plan.md) · [Meeting interpretation](../meeting/README.en.md)

## Reference architecture

1. Capture separate original tracks for participants A and B through your call platform. Call the interpretation API for each direction from your server.
2. Route A's translated speech to B and B's to A. Distribute captions alongside the corresponding direction. Keep original and synthesized tracks separate so the translated audio does not re-enter the input.
3. On hangup, finish both inputs, collect the remaining results, and release resources. For phone support, use a dedicated SIP or media gateway to handle numbering, codecs, and call lifecycle.

Start with the [classroom example](../classroom/README.en.md) to verify one input and output, then use two authorized recordings to check both translation directions. In a live call, separately verify overlapping speech, echo, mute, hangup finalization, and concurrency limits. [Interpretation v3 reference](https://doc.vuilabs.cn/api-reference/realtime-interpretation/)
