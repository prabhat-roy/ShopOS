import { useEffect, useState } from 'react'
import StatCard from '../components/StatCard'
import { statsApi } from '../api/client'
import type { DashboardStats } from '../types'

export default function Dashboard() {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  useEffect(() => { statsApi.get().then(setStats).catch(() => {}) }, [])

  return (
    <div>
      <h1 style={{fontSize:'1.5rem',fontWeight:700,marginBottom:'1.5rem'}}>Dashboard</h1>
      <div style={{display:'grid',gridTemplateColumns:'repeat(auto-fill,minmax(220px,1fr))',gap:'1rem',marginBottom:'2rem'}}>
        <StatCard label="Total Revenue"   value={stats?.totalRevenue ?? 0}   change={stats?.revenueChange} prefix="$" />
        <StatCard label="Total Orders"    value={stats?.totalOrders ?? 0}    change={stats?.ordersChange} />
        <StatCard label="Total Users"     value={stats?.totalUsers ?? 0}     change={stats?.usersChange} />
        <StatCard label="Total Products"  value={stats?.totalProducts ?? 0} />
      </div>
      <div style={{background:'#fff',border:'1px solid #e5e7eb',borderRadius:'0.5rem',padding:'1.5rem'}}>
        <h2 style={{fontWeight:600,marginBottom:'1rem'}}>Quick Actions</h2>
        <div style={{display:'flex',gap:'0.75rem',flexWrap:'wrap'}}>
          <a href="/orders"   style={{padding:'0.5rem 1rem',background:'#f3f4f6',borderRadius:'0.375rem',fontSize:'0.875rem'}}>View Orders</a>
          <a href="/products" style={{padding:'0.5rem 1rem',background:'#f3f4f6',borderRadius:'0.375rem',fontSize:'0.875rem'}}>Manage Products</a>
          <a href="/users"    style={{padding:'0.5rem 1rem',background:'#f3f4f6',borderRadius:'0.375rem',fontSize:'0.875rem'}}>View Users</a>
        </div>
      </div>
    </div>
  )
}
