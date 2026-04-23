defmodule PubsubRouter.Router do
  use Plug.Router

  plug :match
  plug Plug.Parsers, parsers: [:json], json_decoder: Jason
  plug :dispatch

  get "/healthz" do
    conn
    |> put_resp_content_type("application/json")
    |> send_resp(200, Jason.encode!(%{status: "ok"}))
  end

  post "/publish" do
    %{"topic" => topic, "message" => message} = conn.body_params
    PubsubRouter.publish(topic, message)
    conn
    |> put_resp_content_type("application/json")
    |> send_resp(200, Jason.encode!(%{status: "published"}))
  end

  match _ do
    send_resp(conn, 404, "not found")
  end
end
