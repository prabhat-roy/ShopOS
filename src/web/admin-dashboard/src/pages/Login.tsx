import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../store/auth'
import { api } from '../api/client'

export default function Login() {
  const { setAuth } = useAuth()
  const nav = useNavigate()
  const [form, setForm] = useState({ email: '', password: '' })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function submit(e: React.FormEvent) {
    e.preventDefault(); setLoading(true); setError('')
    try {
      const { token } = await api.post<{ token: string }>('/admin/auth/login', form)
      setAuth(token, form.email)
      nav('/')
    } catch { setError('Invalid credentials') } finally { setLoading(false) }
  }

  return (
    <div style={{minHeight:'100vh',display:'flex',alignItems:'center',justifyContent:'center',background:'#f9fafb'}}>
      <div style={{background:'#fff',borderRadius:'0.5rem',border:'1px solid #e5e7eb',padding:'2rem',width:360}}>
        <h1 style={{fontSize:'1.5rem',fontWeight:700,marginBottom:'1.5rem'}}>Admin Login</h1>
        {error && <p style={{color:'#dc2626',marginBottom:'1rem',fontSize:'0.875rem'}}>{error}</p>}
        <form onSubmit={submit}>
          <div style={{marginBottom:'1rem'}}><label style={{display:'block',fontSize:'0.875rem',fontWeight:500,marginBottom:'0.25rem'}}>Email</label><input type="email" required value={form.email} onChange={e => setForm(f => ({...f,email:e.target.value}))} style={{width:'100%',border:'1px solid #d1d5db',borderRadius:'0.375rem',padding:'0.5rem 0.75rem'}} /></div>
          <div style={{marginBottom:'1.5rem'}}><label style={{display:'block',fontSize:'0.875rem',fontWeight:500,marginBottom:'0.25rem'}}>Password</label><input type="password" required value={form.password} onChange={e => setForm(f => ({...f,password:e.target.value}))} style={{width:'100%',border:'1px solid #d1d5db',borderRadius:'0.375rem',padding:'0.5rem 0.75rem'}} /></div>
          <button type="submit" disabled={loading} style={{width:'100%',background:'#111',color:'#fff',border:'none',borderRadius:'0.375rem',padding:'0.75rem',fontWeight:500,cursor:'pointer'}}>{loading ? 'Logging in...' : 'Login'}</button>
        </form>
      </div>
    </div>
  )
}
