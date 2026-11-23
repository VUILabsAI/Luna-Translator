package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

const inputRate = 16000

func readWAV(path string) ([]byte, error) {
	const maxPCMBytes = 14 * 60 * inputRate * 2
	const maxWAVBytes = maxPCMBytes + (1 << 20)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, maxWAVBytes+1))
	if err != nil {
		return nil, err
	}
	if len(content) > maxWAVBytes {
		return nil, errors.New("input WAV is too large for this example")
	}
	if len(content) < 12 || string(content[:4]) != "RIFF" || string(content[8:12]) != "WAVE" || binary.LittleEndian.Uint32(content[4:]) != uint32(len(content)-8) {
		return nil, errors.New("invalid RIFF/WAVE header or file length")
	}
	validFormat := false
	var pcm []byte
	for chunks := content[12:]; len(chunks) > 0; {
		if len(chunks) < 8 {
			return nil, errors.New("truncated WAV chunk header")
		}
		sizeField := binary.LittleEndian.Uint32(chunks[4:])
		if sizeField > uint32(len(chunks)-8) {
			return nil, errors.New("truncated WAV chunk")
		}
		size := int(sizeField)
		end := 8 + size + size%2
		if end > len(chunks) {
			return nil, errors.New("truncated WAV chunk")
		}
		data := chunks[8 : 8+size]
		switch string(chunks[:4]) {
		case "fmt ":
			if validFormat || len(data) < 16 || binary.LittleEndian.Uint16(data) != 1 ||
				binary.LittleEndian.Uint16(data[2:]) != 1 || binary.LittleEndian.Uint32(data[4:]) != inputRate ||
				binary.LittleEndian.Uint32(data[8:]) != inputRate*2 || binary.LittleEndian.Uint16(data[12:]) != 2 ||
				binary.LittleEndian.Uint16(data[14:]) != 16 {
				return nil, errors.New("input must be a 16 kHz mono PCM16 WAV")
			}
			validFormat = true
		case "data":
			if pcm != nil || size == 0 || size%2 != 0 || size > maxPCMBytes {
				return nil, errors.New("WAV must contain one non-empty PCM16 data chunk of at most 14 minutes")
			}
			pcm = data
		}
		chunks = chunks[end:]
	}
	if !validFormat || pcm == nil {
		return nil, errors.New("WAV is missing format or PCM data")
	}
	return pcm, nil
}

func writeWAV(path string, sampleRate int, pcm []byte) error {
	header := []byte("RIFF\x00\x00\x00\x00WAVEfmt \x10\x00\x00\x00\x01\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\x10\x00data\x00\x00\x00\x00")
	binary.LittleEndian.PutUint32(header[4:], uint32(36+len(pcm)))
	binary.LittleEndian.PutUint32(header[24:], uint32(sampleRate))
	binary.LittleEndian.PutUint32(header[28:], uint32(sampleRate*2))
	binary.LittleEndian.PutUint32(header[40:], uint32(len(pcm)))
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	if _, err = file.Write(header); err == nil {
		_, err = file.Write(pcm)
	}
	if err = errors.Join(err, file.Close()); err != nil {
		_ = os.Remove(path)
		return fmt.Errorf("save WAV: %w", err)
	}
	return nil
}
