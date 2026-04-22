'use client'
import { useState } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import Link from 'next/link'
import { useAuth } from '@/lib/store/auth'
import { api } from '@/lib/api'
import type { User } from '@/lib/types'

export default function LoginPage() {
  const { setAuth } = useAuth()
  const router = useRouter()
  const params = useSearchParams()
  const [form, setForm] = useState({ email: '', password: '' })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true); setError('')
    try {
      const { user, token } = await api.post<{ user: User; token: string }>('/auth/login', form)
      setAuth(user, token)
      router.push(params.get('from') ?? '/')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed')
    } finally { setLoading(false) }
  }

  return (
    <div className="container" style={{maxWidth:400}}>
      <h1 className="page-title">Login</h1>
      {error && <div className="alert-error">{error}</div>}
      <form onSubmit={submit}>
        <div className="form-group"><label>Email</label><input type="email" required value={form.email} onChange={e => setForm(f => ({...f, email: e.target.value}))} /></div>
        <div className="form-group"><label>Password</label><input type="password" required value={form.password} onChange={e => setForm(f => ({...f, password: e.target.value}))} /></div>
        <button type="submit" className="btn-primary" disabled={loading}>{loading ? 'Logging in...' : 'Login'}</button>
      </form>
      <p style={{marginTop:'1rem',textAlign:'center',color:'#6b7280'}}>
        No account? <Link href="/register" style={{color:'#111',fontWeight:500}}>Register</Link>
      </p>
    </div>
  )
}
