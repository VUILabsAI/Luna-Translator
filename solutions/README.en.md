# Luna interpretation solutions

[Project home](../README.en.md) · [中文](README.md)

Choose a workflow to see the public example or reference architecture and its verification scope.

| Scenario | Current scope |
|---|---|
| [Classrooms and lectures](classroom/README.en.md) | **Runnable example.** Send the included Mandarin lecture clip to interpretation v3, read English translations, and save the translated speech. |
| [Meeting interpretation](meeting/README.en.md) | **Single-stream community workflow.** A presenter can get captions and translated speech in the Web app. Multi-party RTC is an official product capability covered here as an integration reference. |
| [Livestream interpretation](livestream/README.en.md) | **Mac OBS pilot.** The public repository covers single-stream translation. OBS and live platform verification are tracked separately. |
| [Call interpretation](calls/README.en.md) | **Reference architecture.** The community edition covers face-to-face turn-taking. Third-party track injection and phone or SIP integration require separate implementations. |

## Choose an integration path

- **Run the open-source app.** Start the Web + Go community edition with Docker Compose. Simulation mode runs without a key. Live translation uses the operator's configured provider. The community Luna adapter uses `Authorization: Bearer`. [Deployment guide](../docs/deployment.md) · [Community protocols](../docs/providers.md)
- **Integrate interpretation into your system.** The Go Quickstart connects directly to mainland China interpretation v3 using `X-API-Key`. For Web integrations, keep the key on your server. [Go Quickstart](../examples/quickstart/README.en.md) · [Interpretation v3 reference](https://doc.vuilabs.cn/api-reference/realtime-interpretation/)
