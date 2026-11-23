package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const v3Endpoint = "wss://api.vuilabs.cn/v1/interpret/stream"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	code := run(ctx, os.Args[1:], os.Getenv, os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}

func run(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("Luna interpretation v3", flag.ContinueOnError)
	flags.SetOutput(stderr)
	input := flags.String("input", "", "16 kHz mono PCM16 WAV file")
	output := flags.String("output", "", "new WAV file for translated speech")
	source := flags.String("source", "zh", "v3 source language code")
	target := flags.String("target", "en", "v3 target language code")
	voice := flags.String("voice", "", "v3 preset voice ID; no voice cloning")
	textOnly := flags.Bool("text-only", false, "return translated text without speech")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 || *input == "" || *source == "" || *target == "" ||
		(*textOnly && (*voice != "" || *output != "")) || (!*textOnly && (*voice == "" || *output == "")) {
		fmt.Fprintln(stderr, "required: -input and either -text-only or both -voice and -output")
		flags.PrintDefaults()
		return 2
	}
	apiKey := strings.TrimSpace(getenv("VUILABS_API_KEY"))
	if apiKey == "" {
		fmt.Fprintln(stderr, "set VUILABS_API_KEY to an authorized VUI Labs v3 API key")
		return 2
	}
	if *output != "" {
		if _, err := os.Lstat(*output); err == nil {
			fmt.Fprintln(stderr, "output already exists; choose a new file")
			return 2
		} else if !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintln(stderr, err)
			return 2
		}
	}
	pcm, err := readWAV(*input)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	start := startRequest{Action: "start", Source: *source, Target: *target, EnableTTS: !*textOnly}
	if !*textOnly {
		start.VoiceMode, start.Voice = "preset", *voice
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute+30*time.Second)
	defer cancel()
	result, err := interpret(ctx, websocket.DefaultDialer, v3Endpoint, apiKey, start, pcm, stdout, stderr)
	if err == nil && !*textOnly {
		err = writeWAV(*output, result.sampleRate, result.pcm)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if !*textOnly {
		fmt.Fprintf(stderr, "Saved translated speech to %s\n", *output)
	}
	return 0
}
