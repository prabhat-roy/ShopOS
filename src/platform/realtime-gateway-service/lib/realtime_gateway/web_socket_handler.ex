defmodule RealtimeGateway.WebSocketHandler do
  @behaviour :cowboy_websocket

  def init(req, state) do
    {:cowboy_websocket, req, state}
  end

  def websocket_init(state) do
    Phoenix.PubSub.subscribe(RealtimeGateway.PubSub, "realtime")
    {:ok, state}
  end

  def websocket_handle({:text, msg}, state) do
    {:reply, {:text, msg}, state}
  end

  def websocket_handle(_frame, state) do
    {:ok, state}
  end

  def websocket_info({:broadcast, msg}, state) do
    {:reply, {:text, Jason.encode!(msg)}, state}
  end

  def websocket_info(_info, state) do
    {:ok, state}
  end

  def terminate(_reason, _req, _state) do
    :ok
  end
end
