package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestV3Stream(t *testing.T) {
	for _, scenario := range []struct {
		name     string
		speech   bool
		failure  string
		wantErr  string
		wantText string
	}{
		{name: "text and WAV", speech: true, wantText: "Hello.\nAfter finish.\n"},
		{name: "text only", wantText: "Hello.\nAfter finish.\n"},
		{name: "partial synthesis", speech: true, failure: "tts_failed", wantErr: "incomplete", wantText: "Hello.\nAfter finish.\n"},
		{name: "malformed audio", speech: true, failure: "malformed_audio", wantErr: "invalid float32", wantText: "Hello.\nAfter finish.\n"},
		{name: "quota error before started", failure: "quota_exhausted", wantErr: "quota_exhausted"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			input := bytes.Repeat([]byte{1, 2}, 1840)
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.Header.Get("X-API-Key") != "example-test-key" || request.Header.Get("Authorization") != "" || request.URL.RequestURI() != "/v1/interpret/stream" {
					t.Error("unexpected v3 request path or authentication header")
					writer.WriteHeader(http.StatusUnauthorized)
					return
				}
				upgrader := websocket.Upgrader{}
				conn, err := upgrader.Upgrade(writer, request, nil)
				if err != nil {
					t.Error(err)
					return
				}
				defer conn.Close()
				conn.SetReadDeadline(time.Now().Add(3 * time.Second))
				var start startRequest
				if err := conn.ReadJSON(&start); err != nil || start.Action != "start" || start.Source != "zh" || start.Target != "en" ||
					start.EnableTTS != scenario.speech || (scenario.speech && (start.VoiceMode != "preset" || start.Voice != "example-voice")) {
					t.Errorf("unexpected start: %+v (%v)", start, err)
					return
				}
				_ = conn.WriteJSON(streamEvent{Type: "session_opened", RequestID: "request-example"})
				if scenario.failure == "quota_exhausted" {
					_ = conn.WriteJSON(streamEvent{Type: "error", Code: scenario.failure})
					return
				}
				_ = conn.WriteJSON(streamEvent{Type: "started"})
				var got []byte
				for _, size := range []int{3200, 480} {
					kind, data, err := conn.ReadMessage()
					if err != nil || kind != websocket.BinaryMessage || len(data) != size {
						t.Errorf("expected binary PCM of %d bytes, got %d, %v", size, len(data), err)
						return
					}
					got = append(got, data...)
					if size == 3200 {
						_ = conn.WriteJSON(streamEvent{Type: "translation", Text: "Hello."})
					}
				}
				var finish map[string]string
				if err := conn.ReadJSON(&finish); err != nil || finish["action"] != "finish" || !bytes.Equal(got, input) {
					t.Errorf("invalid input or finish: %v, %v", finish, err)
					return
				}
				_ = conn.WriteJSON(streamEvent{Type: "translation", Text: "After finish."})
				if scenario.speech {
					_ = conn.WriteJSON(streamEvent{Type: "speech_chunk", Audio: &speechInfo{SampleRate: 24000, Samples: 2}})
					frame := binary.LittleEndian.AppendUint32(nil, math.Float32bits(-0.5))
					if scenario.failure != "malformed_audio" {
						frame = binary.LittleEndian.AppendUint32(frame, math.Float32bits(0.5))
					}
					_ = conn.WriteMessage(websocket.BinaryMessage, frame)
					if scenario.failure == "tts_failed" {
						_ = conn.WriteJSON(streamEvent{Type: "speech_done", Code: "tts_failed"})
					}
				}
				_ = conn.WriteJSON(streamEvent{Type: "completed"})
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
			defer cancel()
			start := startRequest{Action: "start", Source: "zh", Target: "en", EnableTTS: scenario.speech}
			if scenario.speech {
				start.VoiceMode, start.Voice = "preset", "example-voice"
			}
			var stdout, stderr bytes.Buffer
			result, err := interpret(ctx, websocket.DefaultDialer, "ws"+strings.TrimPrefix(server.URL, "http")+"/v1/interpret/stream", "example-test-key", start, input, &stdout, &stderr)
			if scenario.wantErr == "" && err != nil {
				t.Fatal(err)
			}
			if scenario.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), scenario.wantErr) || !strings.Contains(err.Error(), "request_id=request-example") {
					t.Fatalf("unexpected error: %v", err)
				}
			}
			if stdout.String() != scenario.wantText || stderr.String() != "request_id=request-example\n" {
				t.Fatalf("unexpected output: stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
			if scenario.speech && scenario.failure == "" {
				if result.sampleRate != 24000 || !bytes.Equal(result.pcm, []byte{0, 192, 0, 64}) {
					t.Fatalf("unexpected audio: %+v", result)
				}
			}
		})
	}
}

func TestV3AuthenticationAndTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, "not authorized", http.StatusUnauthorized)
	}))
	defer server.Close()
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	_, err := interpret(context.Background(), websocket.DefaultDialer, url, "example-test-key", startRequest{}, nil, io.Discard, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "authentication") {
		t.Fatalf("unexpected authentication error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = interpret(ctx, websocket.DefaultDialer, url, "example-test-key", startRequest{}, nil, io.Discard, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("unexpected cancellation: %v", err)
	}
}

func TestWAVAndCLIValidation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.wav")
	want := []byte{1, 2, 3, 4}
	if err := writeWAV(path, inputRate, want); err != nil {
		t.Fatal(err)
	}
	if err := writeWAV(path, inputRate, []byte{9, 9}); err == nil {
		t.Fatal("output must not overwrite input")
	}
	if got, err := readWAV(path); err != nil || !bytes.Equal(got, want) {
		t.Fatalf("readWAV = %v, %v", got, err)
	}
	badWAV := filepath.Join(t.TempDir(), "bad.wav")
	if err := os.WriteFile(badWAV, []byte("not a WAV"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readWAV(badWAV); err == nil {
		t.Fatal("accepted malformed WAV")
	}
	var stderr bytes.Buffer
	code := run(context.Background(), []string{"-input", path, "-text-only"}, func(string) string { return "" }, io.Discard, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), "VUILABS_API_KEY") {
		t.Fatalf("missing credential: code=%d, stderr=%q", code, stderr.String())
	}
}
