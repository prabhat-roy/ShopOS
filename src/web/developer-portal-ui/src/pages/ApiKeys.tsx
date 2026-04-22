import { useEffect, useState } from 'react'
import { api } from '../api/client'
import type { ApiKey } from '../types'

const ALL_SCOPES = ['read:products','write:products','read:orders','write:orders','read:customers']

export default function ApiKeys() {
  const [keys, setKeys] = useState<ApiKey[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [name, setName] = useState('')
  const [scopes, setScopes] = useState<string[]>([])
  const [newKey, setNewKey] = useState<string | null>(null)

  useEffect(() => { api.get<ApiKey[]>('/developer/keys').then(setKeys).catch(()=>{}).finally(()=>setLoading(false)) }, [])

  async function create(e: React.FormEvent) {
    e.preventDefault()
    const { key, apiKey } = await api.post<{ key: string; apiKey: ApiKey }>('/developer/keys', { name, scopes })
    setNewKey(key); setKeys(k => [...k, apiKey]); setShowForm(false); setName(''); setScopes([])
  }

  async function revoke(id: string) {
    if (!confirm('Revoke this key?')) return
    await api.delete(`/developer/keys/${id}`)
    setKeys(k => k.filter(x => x.id !== id))
  }

  return (
    <div>
      <div style={{display:'flex',justifyContent:'space-between',alignItems:'center',marginBottom:'1.5rem'}}>
        <h1 style={{fontSize:'1.5rem',fontWeight:700}}>API Keys</h1>
        <button onClick={() => setShowForm(true)} style={{background:'#111',color:'#fff',border:'none',borderRadius:'0.375rem',padding:'0.5rem 1rem',cursor:'pointer',fontSize:'0.875rem'}}>+ Create Key</button>
      </div>

      {newKey && (
        <div style={{background:'#d1fae5',border:'1px solid #6ee7b7',borderRadius:'0.5rem',padding:'1rem',marginBottom:'1rem'}}>
          <p style={{fontWeight:600,marginBottom:'0.5rem'}}>Save this key — it won't be shown again:</p>
          <code style={{background:'#fff',padding:'0.5rem',borderRadius:'0.25rem',display:'block',wordBreak:'break-all'}}>{newKey}</code>
          <button onClick={() => setNewKey(null)} style={{marginTop:'0.5rem',background:'none',border:'none',cursor:'pointer',color:'#065f46',fontSize:'0.875rem'}}>Dismiss</button>
        </div>
      )}

      {showForm && (
        <div style={{background:'#fff',border:'1px solid #e5e7eb',borderRadius:'0.5rem',padding:'1.5rem',marginBottom:'1.5rem'}}>
          <h2 style={{fontWeight:600,marginBottom:'1rem'}}>New API Key</h2>
          <form onSubmit={create}>
            <div style={{marginBottom:'1rem'}}>
              <label style={{display:'block',fontSize:'0.875rem',fontWeight:500,marginBottom:'0.25rem'}}>Key Name</label>
              <input required value={name} onChange={e=>setName(e.target.value)} placeholder="e.g. Production App" style={{width:'100%',border:'1px solid #d1d5db',borderRadius:'0.375rem',padding:'0.5rem 0.75rem',boxSizing:'border-box' as const}} />
            </div>
            <div style={{marginBottom:'1rem'}}>
              <label style={{display:'block',fontSize:'0.875rem',fontWeight:500,marginBottom:'0.5rem'}}>Scopes</label>
              <div style={{display:'flex',flexWrap:'wrap' as const,gap:'0.5rem'}}>
                {ALL_SCOPES.map(s => (
                  <label key={s} style={{display:'flex',alignItems:'center',gap:'0.25rem',fontSize:'0.875rem',cursor:'pointer'}}>
                    <input type="checkbox" checked={scopes.includes(s)} onChange={e => setScopes(sc => e.target.checked ? [...sc,s] : sc.filter(x=>x!==s))} />
                    {s}
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
        {loading ? <p style={{padding:'2rem',textAlign:'center',color:'#9ca3af'}}>Loading...</p> : keys.length === 0 ? <p style={{padding:'2rem',textAlign:'center',color:'#9ca3af'}}>No API keys yet.</p> : (
          <table style={{width:'100%',borderCollapse:'collapse',fontSize:'0.875rem'}}>
            <thead><tr style={{borderBottom:'2px solid #e5e7eb'}}>{['Name','Prefix','Scopes','Created','Last Used',''].map(h=><th key={h} style={{textAlign:'left',padding:'0.75rem 1rem',fontWeight:600}}>{h}</th>)}</tr></thead>
            <tbody>
              {keys.map(k => (
                <tr key={k.id} style={{borderBottom:'1px solid #f3f4f6'}}>
                  <td style={{padding:'0.75rem 1rem',fontWeight:500}}>{k.name}</td>
                  <td style={{padding:'0.75rem 1rem'}}><code style={{background:'#f3f4f6',padding:'0.2rem 0.4rem',borderRadius:'0.25rem'}}>{k.prefix}…</code></td>
                  <td style={{padding:'0.75rem 1rem',fontSize:'0.75rem',color:'#6b7280'}}>{k.scopes.join(', ')}</td>
                  <td style={{padding:'0.75rem 1rem',color:'#6b7280'}}>{new Date(k.createdAt).toLocaleDateString()}</td>
                  <td style={{padding:'0.75rem 1rem',color:'#6b7280'}}>{k.lastUsed ? new Date(k.lastUsed).toLocaleDateString() : 'Never'}</td>
                  <td style={{padding:'0.75rem 1rem'}}><button onClick={()=>revoke(k.id)} style={{color:'#dc2626',background:'none',border:'none',cursor:'pointer',fontSize:'0.875rem'}}>Revoke</button></td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
