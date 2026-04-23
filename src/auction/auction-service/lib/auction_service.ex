defmodule AuctionService do
  use Application

  def start(_, _) do
    children = [
      {Phoenix.PubSub, name: AuctionService.PubSub},
      {Plug.Cowboy, scheme: :http, plug: AuctionService.Router, options: [port: http_port()]}
    ]

    opts = [strategy: :one_for_one, name: AuctionService.Supervisor]
    Supervisor.start_link(children, opts)
  end

  defp http_port do
    System.get_env("HTTP_PORT", "8211") |> String.to_integer()
  end
end
