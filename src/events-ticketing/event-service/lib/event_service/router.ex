defmodule EventService.Router do
  use Plug.Router

  plug :match
  plug Plug.Parsers, parsers: [:json], json_decoder: Jason
  plug :dispatch

  get "/healthz" do
    conn
    |> put_resp_content_type("application/json")
    |> send_resp(200, Jason.encode!(%{status: "ok"}))
  end

  get "/events" do
    conn
    |> put_resp_content_type("application/json")
    |> send_resp(200, Jason.encode!(%{events: []}))
  end

  match _ do
    send_resp(conn, 404, "not found")
  end
end
