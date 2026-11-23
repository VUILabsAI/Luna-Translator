# Classroom and lecture interpretation

[Solutions index](../README.en.md) · [中文](README.md)

Use the existing Go Quickstart to send one presenter audio track to mainland China interpretation v3 at a real-time pace. Read English translations in the terminal or save the translated speech as WAV. The official product's live sharing can inform a one-to-many design. Student playback and classroom management require a separate integration.

## Run the example

Install Go 1.26.10 or newer. Set `VUILABS_API_KEY` in your local environment to a key with mainland China interpretation v3 access and available quota. Running the example sends the audio to v3. The included [Mandarin lecture WAV](sample-zh.wav) was generated with eSpeak NG as 16 kHz mono PCM16. [Go Quickstart](../../examples/quickstart/README.en.md)

Start from the repository root and confirm that translated text arrives.

```sh
go -C examples/quickstart run . -input ../../solutions/classroom/sample-zh.wav -source zh -target en -text-only
```

Run the following commands from the same directory for translated speech. `michael` is an example preset voice. Use a voice ID listed by the API and choose an output file that does not yet exist. [Voice lookup](https://doc.vuilabs.cn/api-reference/realtime-interpretation/#音色查询)

```sh
mkdir -p output
go -C examples/quickstart run . -input ../../solutions/classroom/sample-zh.wav -source zh -target en -voice michael -output ../../output/classroom-zh-en.wav
```

On success you get translated text, a `request_id` on standard error, and `output/classroom-zh-en.wav` from the second invocation. The clip explains how sound travels through air. The model's exact wording may vary. [Usage and interrupted sessions](https://doc.vuilabs.cn/api-reference/realtime-interpretation/#用量额度与中断)

## Live v3 results

The same lecture clip completed in text-only and preset-voice modes. Both returned identical English text, and the speech run produced a 7.76-second WAV. [English translation](translation-en.txt) · [English speech](translation-en.wav) · [Validation record](validation.md#live-v3-validation)

## Verify without a key

```sh
go -C examples/quickstart test -race ./...
```

The tests use a local mock WebSocket to verify protocol handling and failures without calling a live model. For a full class, keep the v3 key on your server, ingest the presenter track, and distribute the returned results. Plan session rollover and finalization around the API's session length and concurrency limits. [Interpretation v3 reference](https://doc.vuilabs.cn/api-reference/realtime-interpretation/)
