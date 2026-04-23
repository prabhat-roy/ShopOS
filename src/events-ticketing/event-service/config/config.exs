import Config

config :event_service,
  port: String.to_integer(System.get_env("PORT") || "50300")

config :logger,
  level: :info
