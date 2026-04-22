'use client'
import { useEffect, useState } from 'react'
import { api } from '@/lib/api'
import type { Order } from '@/lib/types'

export default function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.get<Order[]>('/orders/me').then(setOrders).catch(() => setOrders([])).finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="container"><p>Loading orders...</p></div>

  return (
    <div className="container" style={{maxWidth:700}}>
      <h1 className="page-title">My Orders</h1>
      {orders.length === 0 ? (
        <div className="empty-state"><h2>No orders yet</h2><p>Your orders will appear here.</p></div>
      ) : (
        orders.map(order => (
          <div key={order.id} className="order-card">
            <div className="order-header">
              <div>
                <strong>#{order.id.slice(0,8).toUpperCase()}</strong>
                <p style={{color:'#6b7280',fontSize:'0.875rem'}}>{new Date(order.createdAt).toLocaleDateString()}</p>
              </div>
              <div style={{textAlign:'right'}}>
                <span className={`order-status status-${order.status}`}>{order.status}</span>
                <p style={{fontWeight:700,marginTop:'0.25rem'}}>${order.total.toFixed(2)}</p>
              </div>
            </div>
            <p style={{color:'#6b7280',fontSize:'0.875rem'}}>{order.items.length} item(s)</p>
          </div>
        ))
      )}
    </div>
  )
}
