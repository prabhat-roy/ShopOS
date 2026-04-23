require "kemal"
require "json"

port = (ENV["PORT"]? || "8218").to_i

get "/healthz" do |env|
  env.response.content_type = "application/json"
  {status: "ok"}.to_json
end

post "/webhooks/content" do |env|
  env.response.content_type = "application/json"
  body = env.request.body.try(&.gets_to_end) || "{}"
  payload = JSON.parse(body)
  {status: "received", event: payload["event"]?}.to_json
end

get "/webhooks" do |env|
  env.response.content_type = "application/json"
  {webhooks: [] of String}.to_json
end

Kemal.config.port = port
Kemal.run
