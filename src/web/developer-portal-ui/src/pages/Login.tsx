import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../store/auth'
import { api } from '../api/client'

export default function Login() {
  const { setAuth } = useAuth(); const nav = useNavigate()
  const [form, setForm] = useState({ email:'', password:'' }); const [loading, setLoading] = useState(false); const [error, setError] = useState('')
  async function submit(e: React.FormEvent) {
    e.preventDefault(); setLoading(true); setError('')
    try { const { token } = await api.post<{token:string}>('/developer/auth/login', form); setAuth(token, form.email); nav('/') }
    catch { setError('Invalid credentials') } finally { setLoading(false) }
  }
  return (
    <div style={{minHeight:'100vh',display:'flex',alignItems:'center',justifyContent:'center',background:'#f9fafb',fontFamily:'system-ui,sans-serif'}}>
      <div style={{background:'#fff',border:'1px solid #e5e7eb',borderRadius:'0.5rem',padding:'2rem',width:360}}>
        <h1 style={{fontSize:'1.5rem',fontWeight:700,marginBottom:'1.5rem'}}>Developer Login</h1>
        {error && <p style={{color:'#dc2626',fontSize:'0.875rem',marginBottom:'1rem'}}>{error}</p>}
        <form onSubmit={submit}>
          {(['email','password'] as const).map(k => (
            <div key={k} style={{marginBottom:'1rem'}}>
              <label style={{display:'block',fontSize:'0.875rem',fontWeight:500,marginBottom:'0.25rem'}}>{k.charAt(0).toUpperCase()+k.slice(1)}</label>
              <input type={k} required value={form[k]} onChange={e=>setForm(f=>({...f,[k]:e.target.value}))} style={{width:'100%',border:'1px solid #d1d5db',borderRadius:'0.375rem',padding:'0.5rem 0.75rem',boxSizing:'border-box' as const}} />
            </div>
          ))}
          <button type="submit" disabled={loading} style={{width:'100%',background:'#111',color:'#fff',border:'none',borderRadius:'0.375rem',padding:'0.75rem',fontWeight:500,cursor:'pointer'}}>{loading?'...':'Login'}</button>
        </form>
      </div>
    </div>
  )
}
