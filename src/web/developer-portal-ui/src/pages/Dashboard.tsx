import { useEffect, useState } from 'react'
import { api } from '../api/client'
import type { AppStats } from '../types'

export default function Dashboard() {
  const [stats, setStats] = useState<AppStats | null>(null)
  useEffect(() => { api.get<AppStats>('/developer/stats').then(setStats).catch(()=>{}) }, [])
  const cards = [
    { label: 'Total Requests',   value: stats?.totalRequests?.toLocaleString() ?? '—' },
    { label: 'Error Rate',       value: stats ? `${(stats.errorRate*100).toFixed(2)}%` : '—' },
    { label: 'Active API Keys',  value: stats?.activeKeys ?? '—' },
    { label: 'Active Webhooks',  value: stats?.activeWebhooks ?? '—' },
  ]
  return (
    <div>
      <h1 style={{fontSize:'1.5rem',fontWeight:700,marginBottom:'1.5rem'}}>Dashboard</h1>
      <div style={{display:'grid',gridTemplateColumns:'repeat(auto-fill,minmax(200px,1fr))',gap:'1rem',marginBottom:'2rem'}}>
        {cards.map(c => (
          <div key={c.label} style={{background:'#fff',border:'1px solid #e5e7eb',borderRadius:'0.5rem',padding:'1.5rem'}}>
            <p style={{fontSize:'0.75rem',color:'#6b7280',textTransform:'uppercase',marginBottom:'0.5rem'}}>{c.label}</p>
            <p style={{fontSize:'1.875rem',fontWeight:700}}>{c.value}</p>
          </div>
        ))}
      </div>
      <div style={{background:'#fff',border:'1px solid #e5e7eb',borderRadius:'0.5rem',padding:'1.5rem'}}>
        <h2 style={{fontWeight:600,marginBottom:'1rem'}}>Getting Started</h2>
        <ol style={{paddingLeft:'1.25rem',color:'#374151',lineHeight:2}}>
          <li>Create an API key in the <a href="/api-keys" style={{color:'#111',fontWeight:500}}>API Keys</a> section</li>
          <li>Test your integration in the <a href="/sandbox" style={{color:'#111',fontWeight:500}}>Sandbox</a></li>
          <li>Set up <a href="/webhooks" style={{color:'#111',fontWeight:500}}>Webhooks</a> to receive real-time events</li>
          <li>Read the <a href="/docs" style={{color:'#111',fontWeight:500}}>API Documentation</a></li>
        </ol>
      </div>
    </div>
  )
}
