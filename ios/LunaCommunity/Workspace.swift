import AVFoundation
import Foundation
import SwiftUI

@MainActor
final class Workspace: ObservableObject {
    enum Phase {
        case disconnected, connecting, ready, starting, streaming, finishing
    }

    @Published var address: String
    @Published var password = ""
    @Published var source = "zh"
    @Published var target = "en"
    @Published var provider = ""
    @Published private(set) var config: WorkspaceConfig?
    @Published private(set) var phase: Phase = .disconnected
    @Published private(set) var original = ""
    @Published private(set) var translation = ""
    @Published private(set) var receivedAudio = false
    @Published private(set) var runningSample = false
    @Published private(set) var finished = false
    @Published private(set) var error: String?

    private var api: CommunityAPI?
    private var token: String?
    private var task: Task<Void, Never>?
    private var socket: URLSessionWebSocketTask?
    private var capture: AsyncThrowingStream<Data, Error>.Continuation?
    private var audio: CommunityAudio?

    init() {
        #if targetEnvironment(simulator)
        address = "http://localhost:8088"
        #else
        address = ""
        #endif
    }

    var busy: Bool { phase == .starting || phase == .streaming || phase == .finishing }
    var canStart: Bool { phase == .ready && !provider.isEmpty && source != target }

    func connect() {
        guard phase == .disconnected else { return }
        error = nil
        phase = .connecting
        task = Task {
            do {
                let api = try CommunityAPI(address: address)
                do {
                    let (config, token) = try await api.connect(password: password)
                    try Task.checkCancellation()
                    self.api = api
                    self.token = token
                    self.config = config
                    let available = config.providers.filter(\.ready)
                    provider = available.first { $0.id == config.defaultProvider }?.id
                        ?? available.first?.id ?? ""
                    source = config.consumerSourceLanguages.first ?? ""
                    target = config.consumerTargetLanguages.first { $0 != source } ?? ""
                    password = ""
                    phase = .ready
                } catch {
                    api.close()
                    throw error
                }
            } catch {
                if !Task.isCancelled { self.error = error.localizedDescription }
                phase = .disconnected
            }
            task = nil
        }
    }

    func disconnect() {
        guard phase == .ready else { return }
        api?.close()
        api = nil
        token = nil
        config = nil
        original = ""
        translation = ""
        finished = false
        receivedAudio = false
        error = nil
        phase = .disconnected
    }

    func setSource(_ value: String) {
        source = value
        if target == value {
            target = config?.consumerTargetLanguages.first { $0 != value } ?? ""
        }
    }

    func start(sample: Bool = false) {
        guard canStart, let api, let token, !sample || provider == "demo" else { return }
        phase = .starting
        runningSample = sample
        error = nil
        original = ""
        translation = ""
        receivedAudio = false
        finished = false
        task = Task {
            defer {
                audio?.close()
                audio = nil
                socket?.cancel(with: .normalClosure, reason: nil)
                socket = nil
                capture?.finish()
                capture = nil
                runningSample = false
                phase = .ready
                task = nil
            }
            do {
                if !sample, !(await AVAudioApplication.requestRecordPermission()) {
                    throw CommunityError.microphoneDenied
                }
                try Task.checkCancellation()
                let ticket = try await api.prepare(provider: provider, token: token)
                try Task.checkCancellation()
                let socket = try api.stream()
                self.socket = socket
                let frame = StartFrame(
                    sessionId: ticket.sessionId, ticket: ticket.ticket, provider: provider,
                    sourceLanguage: source, targetLanguages: [target]
                )
                try await send(frame, through: socket)
                let firstEvent = try await nextEvent(from: socket)
                if firstEvent.type == "error" { throw serverError(firstEvent) }
                guard firstEvent.type == "session_started" else { throw CommunityError.invalidResponse }

                let audio = CommunityAudio()
                self.audio = audio
                let (chunks, continuation) = AsyncThrowingStream<Data, Error>.makeStream(
                    bufferingPolicy: .bufferingOldest(32)
                )
                if !sample { capture = continuation }
                try audio.start(capture: capture)
                phase = .streaming

                try await withThrowingTaskGroup(of: Bool.self) { group in
                    group.addTask { [self] in
                        if sample {
                            let silence = Data(repeating: 0, count: 3200)
                            for _ in 0..<12 {
                                try Task.checkCancellation()
                                try await socket.send(.data(silence))
                                try await Task.sleep(for: .milliseconds(100))
                            }
                            await sampleSent()
                        } else {
                            for try await chunk in chunks {
                                try await socket.send(.data(chunk))
                            }
                        }
                        try await send(["type": "stop"], through: socket)
                        return false
                    }
                    group.addTask { [self] in
                        try await receive(from: socket, audio: audio)
                        return true
                    }
                    do {
                        while let receiverFinished = try await group.next(), !receiverFinished {}
                    } catch {
                        socket.cancel(with: .goingAway, reason: nil)
                        throw error
                    }
                    group.cancelAll()
                }
                try await audio.drain()
                finished = true
            } catch {
                if !Task.isCancelled { self.error = error.localizedDescription }
            }
        }
    }

    func stop() {
        guard phase == .streaming, let capture else { return }
        phase = .finishing
        audio?.stopCapture()
        capture.finish()
    }

    func cancel() {
        guard busy || phase == .connecting else { return }
        audio?.stopCapture()
        capture?.finish()
        socket?.cancel(with: .goingAway, reason: nil)
        task?.cancel()
    }

    private func sampleSent() {
        if phase == .streaming { phase = .finishing }
    }

    private func serverError(_ event: StreamEvent) -> CommunityError {
        event.error.map(CommunityError.server) ?? .invalidResponse
    }

    private func receive(from socket: URLSessionWebSocketTask, audio: CommunityAudio) async throws {
        var rate: Int?
        while !Task.isCancelled {
            let event = try await nextEvent(from: socket)
            switch event.type {
            case "source_done", "source_update":
                if let text = event.text { original = text }
            case "translation_done", "translation_update":
                if let text = event.text { translation = text }
            case "translation_delta":
                if let text = event.text { translation += text }
            case "audio_config":
                rate = event.sampleRate
            case "audio":
                guard let pcm = event.audio, let rate else { throw CommunityError.invalidResponse }
                try audio.play(pcm, rate: rate)
                receivedAudio = true
            case "state" where event.state == "finished":
                return
            case "error":
                throw serverError(event)
            default:
                break
            }
        }
        throw CancellationError()
    }
}

private struct StartFrame: Encodable {
    let type = "start"
    let sessionId: String
    let ticket: String
    let provider: String
    let sourceLanguage: String
    let targetLanguages: [String]
}

private func send(_ value: some Encodable, through socket: URLSessionWebSocketTask) async throws {
    let data = try JSONEncoder().encode(value)
    try await socket.send(.string(String(decoding: data, as: UTF8.self)))
}

private func nextEvent(from socket: URLSessionWebSocketTask) async throws -> StreamEvent {
    let data: Data
    switch try await socket.receive() {
    case .string(let text): data = Data(text.utf8)
    case .data(let bytes): data = bytes
    @unknown default: throw CommunityError.invalidResponse
    }
    return try JSONDecoder().decode(StreamEvent.self, from: data)
}
