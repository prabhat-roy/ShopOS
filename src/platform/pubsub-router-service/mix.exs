defmodule PubsubRouter.MixProject do
  use Mix.Project

  def project do
    [
      app: :pubsub_router,
      version: "1.0.0",
      elixir: "~> 1.17",
      start_permanent: Mix.env() == :prod,
      deps: deps()
    ]
  end

  def application do
    [
      mod: {PubsubRouter, []},
      extra_applications: [:logger]
    ]
  end

  defp deps do
    [
      {:plug_cowboy, "~> 2.7"},
      {:jason, "~> 1.4"},
      {:phoenix_pubsub, "~> 2.1"}
    ]
  end
end
