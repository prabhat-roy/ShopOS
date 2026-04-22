'use client'
import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { useAuth } from '@/lib/store/auth'
import { api } from '@/lib/api'
import type { User } from '@/lib/types'

export default function RegisterPage() {
  const { setAuth } = useAuth()
  const router = useRouter()
  const [form, setForm] = useState({ firstName:'', lastName:'', email:'', password:'' })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const set = (k: string) => (e: React.ChangeEvent<HTMLInputElement>) => setForm(f => ({ ...f, [k]: e.target.value }))

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true); setError('')
    try {
      const { user, token } = await api.post<{ user: User; token: string }>('/auth/register', form)
      setAuth(user, token)
      router.push('/')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed')
    } finally { setLoading(false) }
  }

  return (
    <div className="container" style={{maxWidth:400}}>
      <h1 className="page-title">Create Account</h1>
      {error && <div className="alert-error">{error}</div>}
      <form onSubmit={submit}>
        <div style={{display:'grid',gridTemplateColumns:'1fr 1fr',gap:'1rem'}}>
          <div className="form-group"><label>First Name</label><input required value={form.firstName} onChange={set('firstName')} /></div>
          <div className="form-group"><label>Last Name</label><input required value={form.lastName} onChange={set('lastName')} /></div>
        </div>
        <div className="form-group"><label>Email</label><input type="email" required value={form.email} onChange={set('email')} /></div>
        <div className="form-group"><label>Password</label><input type="password" required minLength={8} value={form.password} onChange={set('password')} /></div>
        <button type="submit" className="btn-primary" disabled={loading}>{loading ? 'Creating...' : 'Create Account'}</button>
      </form>
      <p style={{marginTop:'1rem',textAlign:'center',color:'#6b7280'}}>
        Already have an account? <Link href="/login" style={{color:'#111',fontWeight:500}}>Login</Link>
      </p>
    </div>
  )
}
