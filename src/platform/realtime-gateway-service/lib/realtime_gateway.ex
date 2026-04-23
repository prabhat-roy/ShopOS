defmodule RealtimeGateway do
  use Application

  def start(_, _) do
    children = [
      {Phoenix.PubSub, name: RealtimeGateway.PubSub},
      {Plug.Cowboy,
       scheme: :http,
       plug: RealtimeGateway.Router,
       options: [
         port: http_port(),
         dispatch: dispatch()
       ]}
    ]

    opts = [strategy: :one_for_one, name: RealtimeGateway.Supervisor]
    Supervisor.start_link(children, opts)
  end

  defp http_port do
    System.get_env("HTTP_PORT", "8213") |> String.to_integer()
  end

  defp dispatch do
    [
      {:_,
       [
         {"/ws", RealtimeGateway.WebSocketHandler, []},
         {:_, Plug.Cowboy.Handler, {RealtimeGateway.Router, []}}
       ]}
    ]
  end
end
