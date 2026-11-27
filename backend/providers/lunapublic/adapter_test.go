package lunapublic

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"simultrans/core"
)

func TestBearerTwoSegments(t *testing.T) {
	first := bytes.Repeat([]byte{8, 1}, 1600)
	pause := make([]byte, 3200)
	second := bytes.Repeat([]byte{16, 2}, 1600)
	inputFrames := [][]byte{first, pause, second}
	audioPayload := make([]byte, 8)
	binary.LittleEndian.PutUint32(audioPayload, math.Float32bits(.5))
	binary.LittleEndian.PutUint32(audioPayload[4:], math.Float32bits(-.5))

	for _, test := range []struct {
		name     string
		vad      bool
		hotwords []Hotword
	}{
		{name: "VAD off by default"},
		{name: "VAD on with hotwords", vad: true, hotwords: []Hotword{{Source: "西湖", Target: "West Lake"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			usage := make(chan map[string]int, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer test-key" {
					t.Error("missing server-side authorization")
				}
				ws, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
				if err != nil {
					t.Error(err)
					return
				}
				defer ws.Close()
				write := func(value any) bool {
					if err := ws.WriteJSON(value); err != nil {
						t.Error(err)
						return false
					}
					return true
				}
				var raw map[string]json.RawMessage
				if err := ws.ReadJSON(&raw); err != nil {
					t.Error(err)
					return
				}
				for key := range raw {
					if !slices.Contains([]string{"action", "source_lang", "target_lang", "enable_tts", "enable_vad", "tts_backend", "voice_mode", "voice", "hotwords"}, key) {
						t.Errorf("unexpected public start field %q", key)
					}
				}
				var start startRequest
				encoded, _ := json.Marshal(raw)
				if err := json.Unmarshal(encoded, &start); err != nil {
					t.Error(err)
					return
				}
				expected := startRequest{
					Action: "start", SourceLang: "zh", TargetLang: "en",
					EnableTTS: true, EnableVAD: test.vad, Hotwords: test.hotwords,
					TTSBackend: "lunasi-tts-sc", VoiceMode: "preset", Voice: "avery",
				}
				if !reflect.DeepEqual(start, expected) {
					t.Errorf("unexpected public start: %+v", start)
				}
				if !write(map[string]string{"type": "started"}) {
					return
				}

				sendSegment := func(index int, text string, start, end int) bool {
					if !write(map[string]any{"type": "translation", "index": index, "text": text}) ||
						!write(map[string]any{
							"type": "speech_chunk", "index": index, "text_start": start, "text_end": end,
							"audio": map[string]any{"encoding": "pcm_f32le", "sample_rate": 24000, "samples": 2},
						}) {
						return false
					}
					if err := ws.WriteMessage(websocket.BinaryMessage, audioPayload); err != nil {
						t.Error(err)
						return false
					}
					return true
				}

				inputSamples := 0
				for i, frame := range inputFrames {
					kind, got, err := ws.ReadMessage()
					if err != nil || kind != websocket.BinaryMessage || !bytes.Equal(frame, got) {
						t.Errorf("missing or altered input frame %d: kind=%d bytes=%d err=%v", i, kind, len(got), err)
						return
					}
					inputSamples += len(got) / 2
					if i == 1 && test.vad && !sendSegment(0, "First.", 0, 6) {
						return
					}
				}
				var finish map[string]string
				if err := ws.ReadJSON(&finish); err != nil || finish["action"] != "finish" {
					t.Errorf("expected finish after both input segments: %v, %v", finish, err)
					return
				}
				if !test.vad && !sendSegment(0, "First.", 0, 6) {
					return
				}
				if !sendSegment(1, " Second.", 6, 14) {
					return
				}
				// A segment may finish after audio for the next segment has arrived.
				if !write(map[string]any{"type": "speech_done", "index": 0}) ||
					!write(map[string]any{"type": "speech_done", "index": 1}) ||
					!write(map[string]string{"type": "translation_complete"}) {
					return
				}
				inputTokens := (inputSamples*7 + 15999) / 16000
				meter := map[string]int{
					"input_audio_samples": inputSamples, "output_audio_samples": 4,
					"input_audio_tokens": inputTokens, "output_text_tokens": 4,
					"output_audio_tokens": 1, "billable_tokens": inputTokens + 5,
				}
				if write(map[string]any{
					"type": "usage", "input_audio_samples": meter["input_audio_samples"],
					"output_audio_samples": meter["output_audio_samples"], "input_audio_tokens": meter["input_audio_tokens"],
					"output_text_tokens": 4, "output_audio_tokens": 1, "billable_tokens": meter["billable_tokens"],
				}) && write(map[string]string{"type": "completed"}) {
					usage <- meter
				}
			}))
			defer server.Close()

			input := make(chan []byte, len(inputFrames))
			for _, frame := range inputFrames {
				input <- frame
			}
			close(input)
			events := make(chan core.Event, 16)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			adapter := Adapter{APIKey: "test-key", Endpoint: "ws" + strings.TrimPrefix(server.URL, "http"), EnableVAD: test.vad, Hotwords: test.hotwords}
			err := adapter.Run(ctx, core.SessionConfig{
				SourceLanguage: "zh", TargetLanguage: "en", Qwen3SimulSystemPrompt: "legacy client prompt",
			}, input, events)
			if err != nil {
				t.Fatal(err)
			}
			close(events)
			var deltas, audioIDs, audioTexts, doneIDs []string
			var finalText string
			var audioFrames int
			for event := range events {
				switch event.Type {
				case "translation_delta":
					deltas = append(deltas, event.Text)
				case "translation_done":
					finalText = event.Text
				case "audio":
					audioFrames++
					audioIDs = append(audioIDs, event.SegmentID)
					audioTexts = append(audioTexts, event.SegmentText)
					if event.Audio != "AEAAwA==" {
						t.Errorf("unexpected PCM16 encoding: %q", event.Audio)
					}
				case "audio_done":
					doneIDs = append(doneIDs, event.SegmentID)
				}
			}
			wantIDs := []string{"luna-speech-0", "luna-speech-1"}
			wantTexts := []string{"First.", " Second."}
			if !slices.Equal(deltas, wantTexts) || finalText != "First. Second." ||
				audioFrames != 2 || !slices.Equal(audioIDs, wantIDs) ||
				!slices.Equal(audioTexts, wantTexts) || !slices.Equal(doneIDs, wantIDs) {
				t.Fatalf("incomplete two-segment result: text=%q final=%q audio IDs=%v audio text=%q done=%v", deltas, finalText, audioIDs, audioTexts, doneIDs)
			}
			select {
			case meter := <-usage:
				want := map[string]int{
					"input_audio_samples": 4800, "output_audio_samples": 4, "input_audio_tokens": 3,
					"output_text_tokens": 4, "output_audio_tokens": 1, "billable_tokens": 8,
				}
				if !reflect.DeepEqual(meter, want) {
					t.Fatalf("mock usage mismatch: %v", meter)
				}
				t.Logf("one start, two translations, two audio segments and mock usage: %v", meter)
			case <-ctx.Done():
				t.Fatal("missing mock usage")
			}
		})
	}
}

func TestBearerCompletion(t *testing.T) {
	for _, test := range []struct {
		name        string
		textOnly    bool
		terminal    map[string]string
		expectedErr string
	}{
		{name: "text only", textOnly: true, terminal: map[string]string{"type": "completed"}},
		{name: "partial synthesis", terminal: map[string]string{"type": "completed", "status": "partial", "code": "tts_failed"}, expectedErr: "tts_failed"},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ws, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
				if err != nil {
					t.Error(err)
					return
				}
				defer ws.Close()
				var start map[string]any
				if err := ws.ReadJSON(&start); err != nil {
					t.Error(err)
					return
				}
				if start["enable_tts"] != !test.textOnly || start["enable_vad"] != false {
					t.Errorf("invalid boolean fields: %v", start)
				}
				if test.textOnly {
					for _, field := range []string{"voice", "voice_mode", "tts_backend", "system_prompt", "hotwords"} {
						if _, exists := start[field]; exists {
							t.Errorf("unexpected text-only field %s", field)
						}
					}
				}
				if err := ws.WriteJSON(map[string]string{"type": "started"}); err != nil {
					t.Error(err)
					return
				}
				var finish map[string]string
				if err := ws.ReadJSON(&finish); err != nil || finish["action"] != "finish" {
					t.Errorf("expected finish: %v %v", finish, err)
					return
				}
				if err := ws.WriteJSON(test.terminal); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()

			input := make(chan []byte)
			close(input)
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			adapter := Adapter{APIKey: "test-key", Endpoint: "ws" + strings.TrimPrefix(server.URL, "http")}
			err := adapter.Run(ctx, core.SessionConfig{SourceLanguage: "zh", TargetLanguage: "en", TextOnly: test.textOnly}, input, make(chan core.Event, 4))
			if test.expectedErr == "" && err != nil {
				t.Fatal(err)
			}
			if test.expectedErr != "" && (err == nil || !strings.Contains(err.Error(), test.expectedErr)) {
				t.Fatalf("expected %s, got %v", test.expectedErr, err)
			}
		})
	}
}

func TestFloatConversionRejectsMalformedAndNaN(t *testing.T) {
	if _, err := Float32ToPCM16([]byte{1}); err == nil {
		t.Fatal("accepted malformed sample")
	}
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, math.Float32bits(float32(math.NaN())))
	if _, err := Float32ToPCM16(b); err == nil {
		t.Fatal("accepted NaN")
	}
}
