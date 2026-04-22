import { useState } from 'react'
import { api } from '../api/client'

const EXAMPLES = [
  { label: 'List Products',  method: 'GET',  path: '/products?limit=5' },
  { label: 'Get Product',    method: 'GET',  path: '/products/{id}' },
  { label: 'Create Order',   method: 'POST', path: '/orders', body: JSON.stringify({ items: [{ productId: 'prod_1', quantity: 1 }] }, null, 2) },
  { label: 'List Orders',    method: 'GET',  path: '/orders' },
]

export default function Sandbox() {
  const [method, setMethod] = useState('GET')
  const [path, setPath] = useState('/products?limit=5')
  const [body, setBody] = useState('')
  const [response, setResponse] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [status, setStatus] = useState<number | null>(null)

  async function send() {
    setLoading(true); setResponse(null)
    try {
      const data = method === 'GET' ? await api.get<unknown>(path) : await api.post<unknown>(path, body ? JSON.parse(body) : {})
      setResponse(JSON.stringify(data, null, 2)); setStatus(200)
    } catch (e) {
      setResponse(e instanceof Error ? e.message : 'Error'); setStatus(500)
    } finally { setLoading(false) }
  }

  return (
    <div>
      <h1 style={{fontSize:'1.5rem',fontWeight:700,marginBottom:'0.5rem'}}>API Sandbox</h1>
      <p style={{color:'#6b7280',fontSize:'0.875rem',marginBottom:'1.5rem'}}>Test API calls directly from your browser using your API key.</p>

      <div style={{marginBottom:'1rem'}}>
        <p style={{fontSize:'0.875rem',fontWeight:500,marginBottom:'0.5rem'}}>Examples</p>
        <div style={{display:'flex',gap:'0.5rem',flexWrap:'wrap' as const}}>
          {EXAMPLES.map(ex => (
            <button key={ex.label} onClick={() => { setMethod(ex.method); setPath(ex.path); setBody(ex.body ?? '') }}
              style={{padding:'0.25rem 0.75rem',border:'1px solid #d1d5db',borderRadius:'0.25rem',background:'#fff',cursor:'pointer',fontSize:'0.75rem'}}>
              {ex.label}
            </button>
          ))}
        </div>
      </div>

      <div style={{background:'#fff',border:'1px solid #e5e7eb',borderRadius:'0.5rem',padding:'1.5rem',marginBottom:'1rem'}}>
        <div style={{display:'flex',gap:'0.5rem',marginBottom:'1rem'}}>
          <select value={method} onChange={e=>setMethod(e.target.value)} style={{border:'1px solid #d1d5db',borderRadius:'0.375rem',padding:'0.5rem',fontFamily:'monospace',fontWeight:600}}>
            <option>GET</option><option>POST</option><option>PUT</option><option>DELETE</option>
          </select>
          <input value={path} onChange={e=>setPath(e.target.value)} style={{flex:1,border:'1px solid #d1d5db',borderRadius:'0.375rem',padding:'0.5rem 0.75rem',fontFamily:'monospace',fontSize:'0.875rem'}} />
          <button onClick={send} disabled={loading} style={{background:'#111',color:'#fff',border:'none',borderRadius:'0.375rem',padding:'0.5rem 1rem',cursor:'pointer',fontWeight:500}}>
            {loading ? '...' : 'Send'}
          </button>
        </div>
        {method !== 'GET' && (
          <div>
            <label style={{display:'block',fontSize:'0.875rem',fontWeight:500,marginBottom:'0.25rem'}}>Request Body (JSON)</label>
            <textarea value={body} onChange={e=>setBody(e.target.value)} rows={6} style={{width:'100%',fontFamily:'monospace',fontSize:'0.8rem',border:'1px solid #d1d5db',borderRadius:'0.375rem',padding:'0.5rem',boxSizing:'border-box' as const}} />
          </div>
        )}
      </div>

      {response !== null && (
        <div style={{background:'#fff',border:'1px solid #e5e7eb',borderRadius:'0.5rem',padding:'1.5rem'}}>
          <div style={{display:'flex',justifyContent:'space-between',marginBottom:'0.75rem'}}>
            <span style={{fontWeight:600,fontSize:'0.875rem'}}>Response</span>
            <span style={{padding:'0.2rem 0.5rem',borderRadius:'0.25rem',fontSize:'0.75rem',fontWeight:600,background:status===200?'#d1fae5':'#fee2e2',color:status===200?'#065f46':'#dc2626'}}>{status}</span>
          </div>
          <pre style={{fontFamily:'monospace',fontSize:'0.8rem',overflow:'auto',maxHeight:300,margin:0}}>{response}</pre>
        </div>
      )}
    </div>
  )
}
