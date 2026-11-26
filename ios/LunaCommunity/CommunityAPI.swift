import Foundation

struct Language: Identifiable, Hashable {
    let id: String

    var label: String {
        Locale.current.localizedString(forLanguageCode: id) ?? id
    }
}

enum CommunityError: LocalizedError {
    case invalidAddress
    case incompatibleServer
    case invalidResponse
    case server(String)
    case microphoneDenied
    case audioUnavailable
    case audioBacklog

    var errorDescription: String? {
        switch self {
        case .invalidAddress: String(localized: "Enter a workspace HTTPS address.")
        case .incompatibleServer: String(localized: "This server does not support the community protocol.")
        case .invalidResponse: String(localized: "The server returned an unexpected response.")
        case .server(let value): value
        case .microphoneDenied: String(localized: "Allow microphone access in Settings, then try again.")
        case .audioUnavailable: String(localized: "Audio is unavailable. Check the input and output devices.")
        case .audioBacklog: String(localized: "The connection cannot keep up with the microphone.")
        }
    }
}

struct WorkspaceConfig: Decodable {
    struct Provider: Decodable, Identifiable {
        let id: String
        let ready: Bool

        var label: String {
            switch id {
            case "demo": String(localized: "Simulation")
            case "luna": "Luna"
            case "aliyun": "Qwen LiveTranslate"
            default: id
            }
        }
    }

    let community: Bool?
    let eventSchemaVersion: Int?
    let defaultProvider: String
    let providers: [Provider]
    let consumerSourceLanguages: [String]
    let consumerTargetLanguages: [String]
}

struct StreamTicket: Decodable {
    let sessionId: String
    let provider: String
    let ticket: String
    let eventSchemaVersion: Int
}

struct StreamEvent: Decodable {
    let type: String
    let state: String?
    let text: String?
    let audio: Data?
    let sampleRate: Int?
    let error: String?
}

private final class WorkspaceRedirectPolicy: NSObject, URLSessionTaskDelegate {
    func urlSession(
        _ session: URLSession,
        task: URLSessionTask,
        willPerformHTTPRedirection response: HTTPURLResponse,
        newRequest request: URLRequest
    ) async -> URLRequest? {
        // A different endpoint must be entered explicitly before sending a workspace password.
        nil
    }
}

struct CommunityAPI {
    private struct Login: Decodable { let accessToken: String }
    private struct Problem: Decodable { let error: String }

    private let base: URL
    private let network: URLSession

    init(address: String) throws {
        guard let parts = URLComponents(string: address.trimmingCharacters(in: .whitespacesAndNewlines)),
              let host = parts.host, !host.isEmpty, parts.user == nil, parts.password == nil,
              parts.query == nil, parts.fragment == nil, let url = parts.url else {
            throw CommunityError.invalidAddress
        }
        let loopback = ["localhost", "127.0.0.1"].contains(host.lowercased())
        #if targetEnvironment(simulator)
        let allowLocalHTTP = loopback
        #else
        let allowLocalHTTP = false
        #endif
        guard parts.scheme == "https" || (parts.scheme == "http" && allowLocalHTTP) else {
            throw CommunityError.invalidAddress
        }
        base = url
        let config = URLSessionConfiguration.ephemeral
        config.timeoutIntervalForRequest = 12
        config.httpShouldSetCookies = false
        network = URLSession(configuration: config, delegate: WorkspaceRedirectPolicy(), delegateQueue: nil)
    }

    func close() { network.invalidateAndCancel() }

    func connect(password: String) async throws -> (WorkspaceConfig, String) {
        let config: WorkspaceConfig = try await request("api/config")
        guard config.community == true, config.eventSchemaVersion == 1 else {
            throw CommunityError.incompatibleServer
        }
        let login: Login = try await request("api/community/login", fields: ["password": password])
        return (config, login.accessToken)
    }

    func prepare(provider: String, token: String) async throws -> StreamTicket {
        let ticket: StreamTicket = try await request(
            "api/consumer/ws/session", fields: ["provider": provider], token: token
        )
        guard ticket.eventSchemaVersion == 1, ticket.provider == provider else {
            throw CommunityError.incompatibleServer
        }
        return ticket
    }

    func stream() throws -> URLSessionWebSocketTask {
        let endpoint = base.appending(path: "ws/translate")
        guard var parts = URLComponents(url: endpoint, resolvingAgainstBaseURL: false) else {
            throw CommunityError.invalidAddress
        }
        parts.scheme = parts.scheme == "https" ? "wss" : "ws"
        guard let url = parts.url else { throw CommunityError.invalidAddress }
        let socket = network.webSocketTask(with: url)
        socket.maximumMessageSize = 1 << 20
        socket.resume()
        return socket
    }

    private func request<T: Decodable>(
        _ path: String, fields: [String: String]? = nil, token: String? = nil
    ) async throws -> T {
        var request = URLRequest(url: base.appending(path: path))
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        if let token { request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization") }
        if let fields {
            request.httpMethod = "POST"
            request.setValue("application/json", forHTTPHeaderField: "Content-Type")
            request.httpBody = try JSONEncoder().encode(fields)
        }
        let (data, response) = try await network.data(for: request)
        let decoder = JSONDecoder()
        guard let response = response as? HTTPURLResponse else { throw CommunityError.invalidResponse }
        guard (200..<300).contains(response.statusCode) else {
            if let problem = try? decoder.decode(Problem.self, from: data) {
                throw CommunityError.server(problem.error)
            }
            throw CommunityError.invalidResponse
        }
        return try decoder.decode(T.self, from: data)
    }
}
