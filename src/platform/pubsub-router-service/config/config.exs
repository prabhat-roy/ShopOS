import Config

config :pubsub_router,
  port: String.to_integer(System.get_env("PORT") || "50352")

config :logger,
  level: :info
