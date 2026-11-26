import XCTest

final class WorkspaceTests: XCTestCase {
    @MainActor
    func testTwoSimulatedSessionsAndAudio() {
        let app = connect()
        let sample = app.buttons["session-demo"]
        for _ in 0..<2 {
            XCTAssertTrue(sample.waitForExistence(timeout: 15))
            if !sample.isHittable { app.swipeUp() }
            sample.tap()
            assertDemoCompleted(app)
        }
        let screenshot = XCTAttachment(screenshot: app.screenshot())
        screenshot.name = "Luna iOS community simulation"
        screenshot.lifetime = .keepAlways
        add(screenshot)
    }

    @MainActor
    func testMicrophoneStartAndStopWithSimulationProvider() {
        let app = connect()
        let start = app.buttons["session-start"]
        XCTAssertTrue(start.waitForExistence(timeout: 10))
        if !start.isHittable { app.swipeUp() }
        start.tap()

        let system = XCUIApplication(bundleIdentifier: "com.apple.springboard")
        let permission = system.alerts.buttons["Allow"]
        if permission.waitForExistence(timeout: 3) { permission.tap() }

        let stop = app.buttons["session-stop"]
        XCTAssertTrue(stop.waitForExistence(timeout: 10))
        XCTAssertTrue(app.staticTexts["translated-caption"].waitForExistence(timeout: 10))
        if !stop.isHittable { app.swipeUp() }
        stop.tap()
        assertDemoCompleted(app)
    }

    @MainActor
    private func connect() -> XCUIApplication {
        let app = XCUIApplication()
        app.launchArguments += ["-AppleLanguages", "(en)", "-AppleLocale", "en_US"]
        app.launch()

        let password = app.secureTextFields["workspace-password"]
        XCTAssertTrue(password.waitForExistence(timeout: 10))
        password.tap()
        password.typeText("luna-local-demo")
        app.buttons["workspace-connect"].tap()
        return app
    }

    @MainActor
    private func assertDemoCompleted(_ app: XCUIApplication) {
        let completion = app.staticTexts.containing(
            NSPredicate(format: "label CONTAINS %@", "Audio received")
        ).firstMatch
        XCTAssertTrue(completion.waitForExistence(timeout: 15))
        XCTAssertEqual(app.staticTexts["original-caption"].label, "这是模拟会话，不识别实际语音。")
        XCTAssertEqual(
            app.staticTexts["translated-caption"].label,
            "This is a simulated session, not a translation of your speech."
        )
    }
}
