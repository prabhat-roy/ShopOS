'use client'
import { useEffect, useCallback } from 'react'

type NotificationEvent = {
  type: 'order_status' | 'price_drop' | 'back_in_stock' | 'promo'
  payload: Record<string, unknown>
}

type Handler = (event: NotificationEvent) => void

const WS_URL = process.env.NEXT_PUBLIC_WS_URL ?? 'ws://localhost:50132'

export function useRealtimeNotifications(userId: string | undefined, onEvent: Handler) {
  const stableHandler = useCallback(onEvent, [onEvent])

  useEffect(() => {
    if (!userId) return

    let ws: WebSocket
    let reconnectTimer: ReturnType<typeof setTimeout>
    let closed = false

    function connect() {
      ws = new WebSocket(`${WS_URL}/ws/notifications?userId=${userId}`)

      ws.onmessage = (e) => {
        try { stableHandler(JSON.parse(e.data) as NotificationEvent) } catch {}
      }

      ws.onclose = () => {
        if (!closed) reconnectTimer = setTimeout(connect, 5000)
      }

      ws.onerror = () => ws.close()
    }

    connect()
    return () => {
      closed = true
      clearTimeout(reconnectTimer)
      ws?.close()
    }
  }, [userId, stableHandler])
}
