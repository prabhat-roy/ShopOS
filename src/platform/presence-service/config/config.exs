import Config

config :presence_service,
  http_port: String.to_integer(System.get_env("HTTP_PORT") || "8212"),
  grpc_port: String.to_integer(System.get_env("PORT") || "50350")

config :logger,
  level: :info
