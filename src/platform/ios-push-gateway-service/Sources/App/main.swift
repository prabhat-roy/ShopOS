import Vapor

var env = try Environment.detect()
try LoggingSystem.bootstrap(from: &env)
let app = Application(env)
defer { app.shutdown() }

let port = Int(Environment.get("PORT") ?? "8214") ?? 8214
app.http.server.configuration.port = port

app.get("healthz") { _ -> [String: String] in
    return ["status": "ok"]
}

app.post("push") { req -> [String: String] in
    return ["status": "queued"]
}

try app.run()
