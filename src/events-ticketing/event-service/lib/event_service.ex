defmodule EventService do
  use Application

  def start(_, _) do
    children = [
      {Plug.Cowboy, scheme: :http, plug: EventService.Router, options: [port: port()]}
    ]

    opts = [strategy: :one_for_one, name: EventService.Supervisor]
    Supervisor.start_link(children, opts)
  end

  defp port do
    System.get_env("PORT", "50300") |> String.to_integer()
  end
end
