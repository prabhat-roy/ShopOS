import { useEffect, useState } from 'react'
import { api } from '../api/client'
import type { ApiUsage } from '../types'

export default function Analytics() {
  const [usage, setUsage] = useState<ApiUsage[]>([])
  const [loading, setLoading] = useState(true)
  useEffect(() => { api.get<ApiUsage[]>('/developer/usage').then(setUsage).catch(()=>{}).finally(()=>setLoading(false)) }, [])
  return (
    <div>
      <h1 style={{fontSize:'1.5rem',fontWeight:700,marginBottom:'1.5rem'}}>API Analytics</h1>
      <div style={{background:'#fff',border:'1px solid #e5e7eb',borderRadius:'0.5rem',overflow:'hidden'}}>
        {loading ? <p style={{padding:'2rem',textAlign:'center',color:'#9ca3af'}}>Loading...</p> : usage.length === 0 ? <p style={{padding:'2rem',textAlign:'center',color:'#9ca3af'}}>No usage data yet. Make some API calls first.</p> : (
          <table style={{width:'100%',borderCollapse:'collapse',fontSize:'0.875rem'}}>
            <thead><tr style={{borderBottom:'2px solid #e5e7eb'}}>{['Endpoint','Requests','Errors','P50 (ms)','P99 (ms)'].map(h=><th key={h} style={{textAlign:'left',padding:'0.75rem 1rem',fontWeight:600}}>{h}</th>)}</tr></thead>
            <tbody>
              {usage.map((u,i) => (
                <tr key={i} style={{borderBottom:'1px solid #f3f4f6'}}>
                  <td style={{padding:'0.75rem 1rem',fontFamily:'monospace',fontSize:'0.8rem'}}>{u.endpoint}</td>
                  <td style={{padding:'0.75rem 1rem'}}>{u.requests.toLocaleString()}</td>
                  <td style={{padding:'0.75rem 1rem',color:u.errors>0?'#dc2626':'#059669'}}>{u.errors}</td>
                  <td style={{padding:'0.75rem 1rem'}}>{u.p50ms}</td>
                  <td style={{padding:'0.75rem 1rem',color:u.p99ms>500?'#dc2626':'#374151'}}>{u.p99ms}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
