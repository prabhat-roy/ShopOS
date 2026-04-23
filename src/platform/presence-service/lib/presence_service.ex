defmodule PresenceService do
  use Application

  def start(_, _) do
    children = [
      {Phoenix.PubSub, name: PresenceService.PubSub},
      {Plug.Cowboy, scheme: :http, plug: PresenceService.Router, options: [port: http_port()]}
    ]

    opts = [strategy: :one_for_one, name: PresenceService.Supervisor]
    Supervisor.start_link(children, opts)
  end

  defp http_port do
    System.get_env("HTTP_PORT", "8212") |> String.to_integer()
  end
end
