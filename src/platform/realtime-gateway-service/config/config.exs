import Config

config :realtime_gateway,
  http_port: String.to_integer(System.get_env("HTTP_PORT") || "8213"),
  grpc_port: String.to_integer(System.get_env("PORT") || "50351")

config :logger,
  level: :info
