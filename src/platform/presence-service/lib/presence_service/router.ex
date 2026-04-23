defmodule PresenceService.Router do
  use Plug.Router

  plug :match
  plug Plug.Parsers, parsers: [:json], json_decoder: Jason
  plug :dispatch

  get "/healthz" do
    conn
    |> put_resp_content_type("application/json")
    |> send_resp(200, Jason.encode!(%{status: "ok"}))
  end

  post "/presence/:user_id" do
    Phoenix.PubSub.broadcast(PresenceService.PubSub, "presence", {:join, conn.params["user_id"]})
    conn
    |> put_resp_content_type("application/json")
    |> send_resp(200, Jason.encode!(%{status: "online"}))
  end

  delete "/presence/:user_id" do
    Phoenix.PubSub.broadcast(PresenceService.PubSub, "presence", {:leave, conn.params["user_id"]})
    conn
    |> put_resp_content_type("application/json")
    |> send_resp(200, Jason.encode!(%{status: "offline"}))
  end

  match _ do
    send_resp(conn, 404, "not found")
  end
end
