defmodule PresenceService.MixProject do
  use Mix.Project

  def project do
    [
      app: :presence_service,
      version: "1.0.0",
      elixir: "~> 1.17",
      start_permanent: Mix.env() == :prod,
      deps: deps()
    ]
  end

  def application do
    [
      mod: {PresenceService, []},
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
