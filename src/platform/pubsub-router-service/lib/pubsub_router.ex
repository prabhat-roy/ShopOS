defmodule PubsubRouter do
  use Application

  def start(_, _) do
    children = [
      {Phoenix.PubSub, name: PubsubRouter.PubSub},
      {Plug.Cowboy, scheme: :http, plug: PubsubRouter.Router, options: [port: port()]}
    ]

    opts = [strategy: :one_for_one, name: PubsubRouter.Supervisor]
    Supervisor.start_link(children, opts)
  end

  defp port do
    System.get_env("PORT", "50352") |> String.to_integer()
  end

  def publish(topic, message) do
    Phoenix.PubSub.broadcast(__MODULE__.PubSub, topic, message)
  end

  def subscribe(topic) do
    Phoenix.PubSub.subscribe(__MODULE__.PubSub, topic)
  end
end
