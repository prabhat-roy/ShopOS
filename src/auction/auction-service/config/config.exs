import Config

config :auction_service,
  http_port: String.to_integer(System.get_env("HTTP_PORT") || "8211"),
  grpc_port: String.to_integer(System.get_env("PORT") || "50310")

config :logger,
  level: :info
