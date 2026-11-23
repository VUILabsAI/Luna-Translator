package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type startRequest struct {
	Action    string `json:"action"`
	Source    string `json:"source_lang"`
	Target    string `json:"target_lang"`
	EnableTTS bool   `json:"enable_tts"`
	VoiceMode string `json:"voice_mode,omitempty"`
	Voice     string `json:"voice,omitempty"`
}

type speechInfo struct {
	SampleRate int `json:"sample_rate"`
	Samples    int `json:"samples"`
}

type streamEvent struct {
	Type      string      `json:"type"`
	RequestID string      `json:"request_id"`
	Code      string      `json:"code"`
	Text      string      `json:"text"`
	Audio     *speechInfo `json:"audio"`
}

type translatedAudio struct {
	sampleRate int
	pcm        []byte
}

func interpret(ctx context.Context, dialer *websocket.Dialer, endpoint, key string, start startRequest, input []byte, stdout, stderr io.Writer) (_ translatedAudio, err error) {
	conn, response, err := dialer.DialContext(ctx, endpoint, http.Header{"X-Api-Key": {key}})
	if err != nil {
		if response != nil {
			defer response.Body.Close()
			switch response.StatusCode {
			case http.StatusUnauthorized, http.StatusForbidden:
				return translatedAudio{}, fmt.Errorf("v3 authentication or permission denied (HTTP %d); check your API key", response.StatusCode)
			case http.StatusTooManyRequests:
				return translatedAudio{}, errors.New("v3 rate limited (HTTP 429); check quota and concurrency")
			}
			return translatedAudio{}, fmt.Errorf("v3 WebSocket upgrade failed (HTTP %d)", response.StatusCode)
		}
		if ctx.Err() != nil {
			return translatedAudio{}, ctx.Err()
		}
		return translatedAudio{}, fmt.Errorf("connect to v3: %w", err)
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer conn.Close()
	stopClose := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stopClose()
	conn.SetReadLimit(4 << 20)
	var requestID string
	defer func() {
		if err != nil && requestID != "" {
			err = fmt.Errorf("request_id=%s: %w", requestID, err)
		}
	}()
	if err := writeJSON(conn, start); err != nil {
		return translatedAudio{}, err
	}
	go keepAlive(ctx, conn)
	sendErr := make(chan error, 1)
	started, speechFailed := false, false
	var pending *speechInfo
	var result translatedAudio
	for {
		kind, payload, readErr := conn.ReadMessage()
		if readErr != nil {
			select {
			case err := <-sendErr:
				return result, fmt.Errorf("send v3 audio: %w", err)
			default:
			}
			if err := ctx.Err(); err != nil {
				return result, err
			}
			return result, fmt.Errorf("v3 stream closed before completed: %w", readErr)
		}
		if pending != nil {
			if kind != websocket.BinaryMessage {
				return result, errors.New("expected binary audio immediately after speech metadata")
			}
			if err := result.appendSpeech(*pending, payload); err != nil {
				return result, err
			}
			pending = nil
			continue
		}
		if kind != websocket.TextMessage {
			return result, errors.New("unexpected binary frame without speech metadata")
		}
		var event streamEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			return result, fmt.Errorf("decode v3 event: %w", err)
		}
		if requestID == "" && event.RequestID != "" {
			requestID = event.RequestID
			fmt.Fprintf(stderr, "request_id=%s\n", requestID)
		}
		switch event.Type {
		case "started":
			if started {
				return result, errors.New("duplicate started event")
			}
			started = true
			go func() {
				if err := sendAudio(ctx, conn, input); err != nil {
					sendErr <- err
					_ = conn.Close()
				}
			}()
		case "translation":
			if _, err := fmt.Fprintln(stdout, event.Text); err != nil {
				return result, err
			}
		case "speech", "speech_chunk":
			if event.Audio == nil {
				return result, errors.New("missing speech audio metadata")
			}
			pending = event.Audio
		case "tts_error":
			speechFailed = true
		case "speech_done":
			speechFailed = speechFailed || event.Code == "tts_failed"
		case "error":
			if event.Code == "" {
				return result, errors.New("v3 returned an unspecified API error")
			}
			return result, fmt.Errorf("v3 API error: %s", event.Code)
		case "completed":
			if !started {
				return result, errors.New("v3 completed before started")
			}
			if speechFailed || (start.EnableTTS && len(result.pcm) == 0) {
				return result, errors.New("translated speech is incomplete; no WAV saved")
			}
			return result, nil
		}
	}
}

func writeJSON(conn *websocket.Conn, data any) error {
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return conn.WriteJSON(data)
}

func keepAlive(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
				_ = conn.Close()
				return
			}
		}
	}
}

func sendAudio(ctx context.Context, conn *websocket.Conn, input []byte) error {
	const maxFrameSize = inputRate * 2 / 10
	for len(input) > 0 {
		frameSize := min(maxFrameSize, len(input))
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := conn.WriteMessage(websocket.BinaryMessage, input[:frameSize]); err != nil {
			return err
		}
		input = input[frameSize:]
		timer := time.NewTimer(time.Duration(frameSize) * time.Second / (inputRate * 2))
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return writeJSON(conn, map[string]string{"action": "finish"})
}

func (audio *translatedAudio) appendSpeech(info speechInfo, payload []byte) error {
	const maxOutputBytes = 64 << 20
	if (info.SampleRate != 16000 && info.SampleRate != 24000) || info.Samples <= 0 || len(payload)%4 != 0 || len(payload)/4 != info.Samples {
		return errors.New("invalid float32 audio frame or sample rate")
	}
	if audio.sampleRate != 0 && audio.sampleRate != info.SampleRate {
		return errors.New("sample rate changed during session; cannot save a single WAV")
	}
	if len(audio.pcm)+2*info.Samples > maxOutputBytes {
		return errors.New("translated audio exceeds the quickstart WAV size limit")
	}
	audio.sampleRate = info.SampleRate
	for offset := 0; offset < len(payload); offset += 4 {
		sample := float64(math.Float32frombits(binary.LittleEndian.Uint32(payload[offset:])))
		if math.IsNaN(sample) || math.IsInf(sample, 0) {
			return errors.New("non-finite float32 audio sample")
		}
		pcm := int16(math.Round(max(-32768, min(32767, sample*32768))))
		audio.pcm = binary.LittleEndian.AppendUint16(audio.pcm, uint16(pcm))
	}
	return nil
}
