defmodule AuctionService.Router do
  use Plug.Router

  plug :match
  plug Plug.Parsers, parsers: [:json], json_decoder: Jason
  plug :dispatch

  get "/healthz" do
    conn
    |> put_resp_content_type("application/json")
    |> send_resp(200, Jason.encode!(%{status: "ok"}))
  end

  get "/auctions" do
    conn
    |> put_resp_content_type("application/json")
    |> send_resp(200, Jason.encode!(%{auctions: []}))
  end

  post "/auctions/:id/bid" do
    Phoenix.PubSub.broadcast(AuctionService.PubSub, "auction:#{conn.params["id"]}", {:bid, conn.body_params})
    conn
    |> put_resp_content_type("application/json")
    |> send_resp(200, Jason.encode!(%{status: "bid_received"}))
  end

  match _ do
    send_resp(conn, 404, "not found")
  end
end
