import SwiftUI

@main
struct LunaCommunityApp: App {
    @Environment(\.scenePhase) private var scenePhase
    @StateObject private var workspace = Workspace()

    var body: some Scene {
        WindowGroup {
            CommunityScreen(workspace: workspace)
                .onChange(of: scenePhase) { _, phase in
                    if phase == .background { workspace.cancel() }
                }
        }
    }
}

struct CommunityScreen: View {
    @ObservedObject var workspace: Workspace

    private let brand = Color(red: 0.20, green: 0.58, blue: 0.44)

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 20) {
                    if let config = workspace.config {
                        quickStart(config)
                    } else {
                        connection
                    }
                    if let error = workspace.error {
                        Label(error, systemImage: "exclamationmark.circle")
                            .foregroundStyle(.red)
                            .accessibilityIdentifier("workspace-error")
                    }
                }
                .frame(maxWidth: 580)
                .padding(20)
                .frame(maxWidth: .infinity)
            }
            .background(Color(uiColor: .systemGroupedBackground))
            .navigationTitle("Luna Translator")
            .toolbar {
                if workspace.config != nil {
                    ToolbarItem(placement: .topBarTrailing) {
                        Button("Disconnect", action: workspace.disconnect)
                            .disabled(workspace.busy)
                    }
                }
            }
        }
        .tint(brand)
    }

    private var connection: some View {
        VStack(alignment: .leading, spacing: 14) {
            Text("iOS Community Edition").font(.title2.bold())
            Text("Connect to a self-hosted Luna Translator workspace.")
                .foregroundStyle(.secondary)
            VStack(alignment: .leading, spacing: 8) {
                Text("Workspace address").font(.subheadline.weight(.semibold))
                TextField("https://your-workspace.example", text: $workspace.address)
                    .keyboardType(.URL)
                    .textContentType(.URL)
                    .textInputAutocapitalization(.never)
                    .autocorrectionDisabled()
                    .textFieldStyle(.roundedBorder)
                    .accessibilityIdentifier("workspace-address")
                Text("Workspace password").font(.subheadline.weight(.semibold))
                SecureField("Workspace password", text: $workspace.password)
                    .textContentType(.password)
                    .textFieldStyle(.roundedBorder)
                    .accessibilityIdentifier("workspace-password")
            }
            Button(action: workspace.connect) {
                Text(workspace.phase == .connecting ? "Connecting…" : "Connect")
                    .frame(maxWidth: .infinity, minHeight: 32)
            }
            .buttonStyle(.borderedProminent)
            .disabled(workspace.password.isEmpty || workspace.phase == .connecting)
            .accessibilityIdentifier("workspace-connect")
        }
        .padding(16)
        .background(.background, in: RoundedRectangle(cornerRadius: 20))
    }

    private func quickStart(_ config: WorkspaceConfig) -> some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("Quick interpreting").font(.title2.bold())
            VStack {
                selectionRow("Provider") {
                    Picker("Provider", selection: $workspace.provider) {
                        if workspace.provider.isEmpty { Text("No configured provider").tag("") }
                        ForEach(config.providers.filter(\.ready)) { Text($0.label).tag($0.id) }
                    }
                }
                selectionRow("Input language") {
                    Picker("Input language", selection: Binding(get: { workspace.source }, set: workspace.setSource)) {
                        ForEach(config.consumerSourceLanguages.map(Language.init)) { Text($0.label).tag($0.id) }
                    }
                }
                selectionRow("Output language") {
                    Picker("Output language", selection: $workspace.target) {
                        ForEach(config.consumerTargetLanguages.filter { $0 != workspace.source }.map(Language.init)) {
                            Text($0.label).tag($0.id)
                        }
                    }
                }
            }
            .pickerStyle(.menu)
            .frame(maxWidth: .infinity)
            .padding(12)
            .background(.background, in: RoundedRectangle(cornerRadius: 20))
            .disabled(workspace.busy)
            if workspace.provider == "demo" {
                Label("Simulation uses fixed captions and a test tone.", systemImage: "info.circle")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
            }
            caption("Original", text: workspace.original, id: "original-caption")
            caption("Translation", text: workspace.translation, id: "translated-caption")
            if workspace.finished {
                Label(
                    workspace.receivedAudio ? "Session finished · Audio received" : "Session finished",
                    systemImage: "checkmark.circle"
                )
                .font(.footnote)
                .accessibilityIdentifier("session-finished")
            }
            controls
        }
    }

    private func selectionRow<Selection: View>(
        _ title: LocalizedStringKey, @ViewBuilder selection: () -> Selection
    ) -> some View {
        HStack {
            Text(title).foregroundStyle(.secondary).accessibilityHidden(true)
            Spacer()
            selection().labelsHidden()
        }
    }

    private func caption(_ title: LocalizedStringKey, text: String, id: String) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(title).font(.subheadline.weight(.semibold)).foregroundStyle(.secondary)
            if text.isEmpty {
                Text("Captions appear here").foregroundStyle(.secondary)
            } else {
                Text(text).textSelection(.enabled).accessibilityIdentifier(id)
            }
        }
        .frame(maxWidth: .infinity, minHeight: 54, alignment: .topLeading)
        .padding(16)
        .background(.background, in: RoundedRectangle(cornerRadius: 20))
    }

    @ViewBuilder
    private var controls: some View {
        if workspace.busy {
            if workspace.phase == .streaming && !workspace.runningSample {
                Button("End microphone input", action: workspace.stop)
                    .buttonStyle(.borderedProminent)
                    .accessibilityIdentifier("session-stop")
            } else {
                ProgressView(workspace.phase == .finishing ? "Finishing…" : "Starting…")
            }
            Button("Cancel session", action: workspace.cancel)
                .buttonStyle(.bordered)
        } else {
            Button {
                workspace.start()
            } label: {
                Label("Start microphone", systemImage: "mic")
                    .frame(maxWidth: .infinity, minHeight: 32)
            }
            .buttonStyle(.borderedProminent)
            .disabled(!workspace.canStart)
            .accessibilityIdentifier("session-start")
            if workspace.provider == "demo" {
                Button("Run a simulated sample") { workspace.start(sample: true) }
                    .buttonStyle(.bordered)
                    .disabled(!workspace.canStart)
                    .accessibilityIdentifier("session-demo")
            }
        }
    }
}
