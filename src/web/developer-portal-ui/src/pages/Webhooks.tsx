import { useEffect, useState } from 'react'
import { api } from '../api/client'
import type { Webhook } from '../types'

const EVENTS = ['order.created','order.updated','order.fulfilled','payment.processed','payment.failed','product.updated','user.registered']

export default function Webhooks() {
  const [hooks, setHooks] = useState<Webhook[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [url, setUrl] = useState(''); const [events, setEvents] = useState<string[]>([])

  useEffect(() => { api.get<Webhook[]>('/developer/webhooks').then(setHooks).catch(()=>{}).finally(()=>setLoading(false)) }, [])

  async function create(e: React.FormEvent) {
    e.preventDefault()
    const hook = await api.post<Webhook>('/developer/webhooks', { url, events })
    setHooks(h => [...h, hook]); setShowForm(false); setUrl(''); setEvents([])
  }

  async function del(id: string) {
    if (!confirm('Delete this webhook?')) return
    await api.delete(`/developer/webhooks/${id}`)
    setHooks(h => h.filter(x => x.id !== id))
  }

  return (
    <div>
      <div style={{display:'flex',justifyContent:'space-between',alignItems:'center',marginBottom:'1.5rem'}}>
        <h1 style={{fontSize:'1.5rem',fontWeight:700}}>Webhooks</h1>
        <button onClick={()=>setShowForm(true)} style={{background:'#111',color:'#fff',border:'none',borderRadius:'0.375rem',padding:'0.5rem 1rem',cursor:'pointer',fontSize:'0.875rem'}}>+ Add Webhook</button>
      </div>

      {showForm && (
        <div style={{background:'#fff',border:'1px solid #e5e7eb',borderRadius:'0.5rem',padding:'1.5rem',marginBottom:'1.5rem'}}>
          <h2 style={{fontWeight:600,marginBottom:'1rem'}}>New Webhook</h2>
          <form onSubmit={create}>
            <div style={{marginBottom:'1rem'}}>
              <label style={{display:'block',fontSize:'0.875rem',fontWeight:500,marginBottom:'0.25rem'}}>Endpoint URL</label>
              <input type="url" required value={url} onChange={e=>setUrl(e.target.value)} placeholder="https://your-app.com/webhooks" style={{width:'100%',border:'1px solid #d1d5db',borderRadius:'0.375rem',padding:'0.5rem 0.75rem',boxSizing:'border-box' as const}} />
            </div>
            <div style={{marginBottom:'1rem'}}>
              <label style={{display:'block',fontSize:'0.875rem',fontWeight:500,marginBottom:'0.5rem'}}>Events</label>
              <div style={{display:'grid',gridTemplateColumns:'1fr 1fr',gap:'0.25rem'}}>
                {EVENTS.map(ev => (
                  <label key={ev} style={{display:'flex',alignItems:'center',gap:'0.25rem',fontSize:'0.875rem',cursor:'pointer'}}>
                    <input type="checkbox" checked={events.includes(ev)} onChange={e=>setEvents(es=>e.target.checked?[...es,ev]:es.filter(x=>x!==ev))} />{ev}
                  </label>
                ))}
              </div>
            </div>
            <div style={{display:'flex',gap:'0.75rem'}}>
              <button type="submit" style={{background:'#111',color:'#fff',border:'none',borderRadius:'0.375rem',padding:'0.5rem 1rem',cursor:'pointer'}}>Create</button>
              <button type="button" onClick={()=>setShowForm(false)} style={{background:'none',border:'1px solid #d1d5db',borderRadius:'0.375rem',padding:'0.5rem 1rem',cursor:'pointer'}}>Cancel</button>
            </div>
          </form>
        </div>
      )}

      <div style={{background:'#fff',border:'1px solid #e5e7eb',borderRadius:'0.5rem',overflow:'hidden'}}>
        {loading ? <p style={{padding:'2rem',textAlign:'center',color:'#9ca3af'}}>Loading...</p> : hooks.length === 0 ? <p style={{padding:'2rem',textAlign:'center',color:'#9ca3af'}}>No webhooks configured.</p> : (
          <table style={{width:'100%',borderCollapse:'collapse',fontSize:'0.875rem'}}>
            <thead><tr style={{borderBottom:'2px solid #e5e7eb'}}>{['URL','Events','Status','Success Rate',''].map(h=><th key={h} style={{textAlign:'left',padding:'0.75rem 1rem',fontWeight:600}}>{h}</th>)}</tr></thead>
            <tbody>
              {hooks.map(h => (
                <tr key={h.id} style={{borderBottom:'1px solid #f3f4f6'}}>
                  <td style={{padding:'0.75rem 1rem',fontFamily:'monospace',fontSize:'0.8rem'}}>{h.url}</td>
                  <td style={{padding:'0.75rem 1rem',fontSize:'0.75rem',color:'#6b7280'}}>{h.events.length} event(s)</td>
                  <td style={{padding:'0.75rem 1rem'}}><span style={{padding:'0.2rem 0.5rem',borderRadius:'9999px',fontSize:'0.75rem',fontWeight:600,background:h.status==='active'?'#d1fae5':'#fee2e2',color:h.status==='active'?'#065f46':'#dc2626'}}>{h.status}</span></td>
                  <td style={{padding:'0.75rem 1rem'}}>{(h.successRate*100).toFixed(1)}%</td>
                  <td style={{padding:'0.75rem 1rem'}}><button onClick={()=>del(h.id)} style={{color:'#dc2626',background:'none',border:'none',cursor:'pointer',fontSize:'0.875rem'}}>Delete</button></td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
