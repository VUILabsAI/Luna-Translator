import AVFoundation
import Foundation

/// Streaming AVAudioConverter path adapted from the product iOS capture pipeline.
private final class RealtimePCMConverter {
    let format: AVAudioFormat
    private var converter: AVAudioConverter?

    init(format: AVAudioFormat) {
        self.format = format
    }

    func convert(_ input: AVAudioPCMBuffer) -> AVAudioPCMBuffer? {
        guard input.frameLength > 0, input.format.sampleRate > 0 else { return nil }
        if input.format == format { return input }
        if converter?.inputFormat != input.format {
            converter = AVAudioConverter(from: input.format, to: format)
        }
        let ratio = format.sampleRate / input.format.sampleRate
        let capacity = AVAudioFrameCount(Double(input.frameLength) * ratio + 1024)
        guard let converter, let output = AVAudioPCMBuffer(pcmFormat: format, frameCapacity: capacity) else {
            return nil
        }
        var consumed = false
        let status = converter.convert(to: output, error: nil) { _, inputStatus in
            inputStatus.pointee = consumed ? .noDataNow : .haveData
            if consumed { return nil }
            consumed = true
            return input
        }
        guard status != .error, output.frameLength > 0 else { return nil }
        return output
    }
}

@MainActor
final class CommunityAudio {
    private let engine = AVAudioEngine()
    private let player = AVAudioPlayerNode()
    private var playbackConverter: RealtimePCMConverter?
    private var captureInstalled = false
    private var active = false
    private var pendingPlayback = 0

    func start(capture: AsyncThrowingStream<Data, Error>.Continuation?) throws {
        let session = AVAudioSession.sharedInstance()
        do {
            if capture == nil {
                try session.setCategory(.playback, mode: .default)
            } else {
                try session.setCategory(
                    .playAndRecord, mode: .voiceChat, options: [.defaultToSpeaker, .allowBluetoothHFP]
                )
            }
            try session.setActive(true)
            active = true

            // Configure playback before starting capture so first audio doesn't
            // reconnect the live engine graph or invalidate the microphone tap.
            guard let format = AVAudioFormat(standardFormatWithSampleRate: 24_000, channels: 1) else {
                throw CommunityError.audioUnavailable
            }
            playbackConverter = RealtimePCMConverter(format: format)
            engine.attach(player)
            engine.connect(player, to: engine.mainMixerNode, format: format)

            if let capture {
                let input = engine.inputNode
                let inputFormat = input.outputFormat(forBus: 0)
                guard inputFormat.sampleRate > 0, inputFormat.channelCount > 0,
                      let outputFormat = AVAudioFormat(
                        commonFormat: .pcmFormatInt16, sampleRate: 16_000, channels: 1, interleaved: true
                      ) else {
                    throw CommunityError.audioUnavailable
                }
                let converter = RealtimePCMConverter(format: outputFormat)
                input.installTap(onBus: 0, bufferSize: 1024, format: inputFormat) { buffer, _ in
                    guard let output = converter.convert(buffer), let samples = output.int16ChannelData else {
                        return
                    }
                    let bytes = Int(output.frameLength) * MemoryLayout<Int16>.size
                    if case .dropped = capture.yield(Data(bytes: samples[0], count: bytes)) {
                        capture.finish(throwing: CommunityError.audioBacklog)
                    }
                }
                captureInstalled = true
            }
            try engine.start()
        } catch {
            close()
            throw error
        }
    }

    func stopCapture() {
        if captureInstalled {
            engine.inputNode.removeTap(onBus: 0)
            captureInstalled = false
        }
    }

    func play(_ pcm: Data, rate: Int) throws {
        guard let playbackConverter, (8_000...96_000).contains(rate),
              pcm.count.isMultiple(of: 2), !pcm.isEmpty,
              let format = AVAudioFormat(standardFormatWithSampleRate: Double(rate), channels: 1),
              let buffer = AVAudioPCMBuffer(
                pcmFormat: format, frameCapacity: AVAudioFrameCount(pcm.count / 2)
              ), let samples = buffer.floatChannelData else {
            throw CommunityError.invalidResponse
        }
        guard pendingPlayback < 512 else { throw CommunityError.audioBacklog }
        buffer.frameLength = buffer.frameCapacity
        pcm.withUnsafeBytes { bytes in
            for frame in 0..<Int(buffer.frameLength) {
                let value = bytes.loadUnaligned(fromByteOffset: frame * 2, as: Int16.self)
                samples[0][frame] = Float(Int16(littleEndian: value)) / 32768
            }
        }
        guard let output = playbackConverter.convert(buffer) else { return }
        pendingPlayback += 1
        player.scheduleBuffer(output, completionCallbackType: .dataPlayedBack) { [weak self] _ in
            Task { @MainActor in
                guard let self else { return }
                self.pendingPlayback = max(0, self.pendingPlayback - 1)
            }
        }
        if !player.isPlaying { player.play() }
    }

    func drain() async throws {
        let deadline = ContinuousClock.now + .seconds(30)
        while pendingPlayback > 0 {
            guard ContinuousClock.now < deadline else { throw CommunityError.audioUnavailable }
            try await Task.sleep(for: .milliseconds(20))
        }
    }

    func close() {
        stopCapture()
        player.stop()
        engine.stop()
        playbackConverter = nil
        pendingPlayback = 0
        if active {
            try? AVAudioSession.sharedInstance().setActive(false, options: .notifyOthersOnDeactivation)
            active = false
        }
    }
}
