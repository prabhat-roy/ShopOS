import { useEffect, useState } from 'react'
import DataTable from '../components/DataTable'
import { ordersApi } from '../api/client'
import type { Order } from '../types'

const STATUS_COLORS: Record<string,string> = { pending:'#fef3c7', confirmed:'#dbeafe', shipped:'#d1fae5', delivered:'#d1fae5', cancelled:'#fee2e2' }

export default function Orders() {
  const [orders, setOrders] = useState<Order[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => { ordersApi.list().then(setOrders).catch(()=>{}).finally(()=>setLoading(false)) }, [])

  async function updateStatus(id: string, status: string) {
    await ordersApi.update(id, status)
    setOrders(o => o.map(x => x.id === id ? {...x, status} : x))
  }

  return (
    <div>
      <h1 style={{fontSize:'1.5rem',fontWeight:700,marginBottom:'1.5rem'}}>Orders</h1>
      <div style={{background:'#fff',border:'1px solid #e5e7eb',borderRadius:'0.5rem',overflow:'hidden'}}>
        {loading ? <p style={{padding:'2rem',textAlign:'center',color:'#9ca3af'}}>Loading...</p> : (
          <DataTable<Order>
            keyField="id"
            data={orders}
            columns={[
              { key:'id',            header:'Order ID',  render: v => `#${String(v).slice(0,8).toUpperCase()}` },
              { key:'customerEmail', header:'Customer' },
              { key:'itemCount',     header:'Items' },
              { key:'total',         header:'Total',     render: v => `$${Number(v).toFixed(2)}` },
              { key:'createdAt',     header:'Date',      render: v => new Date(String(v)).toLocaleDateString() },
              { key:'status',        header:'Status',    render: (v, row) => (
                <select value={String(v)} onChange={e => updateStatus(row.id, e.target.value)}
                  style={{background: STATUS_COLORS[String(v)] ?? '#f3f4f6',border:'none',borderRadius:'0.25rem',padding:'0.25rem 0.5rem',fontSize:'0.75rem',fontWeight:600,cursor:'pointer'}}>
                  {['pending','confirmed','shipped','delivered','cancelled'].map(s => <option key={s} value={s}>{s}</option>)}
                </select>
              )},
            ]}
          />
        )}
      </div>
    </div>
  )
}
