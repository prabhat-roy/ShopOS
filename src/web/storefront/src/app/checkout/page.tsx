'use client'
import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useCart } from '@/lib/store/cart'
import { api } from '@/lib/api'

export default function CheckoutPage() {
  const { items, total, clear } = useCart()
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [form, setForm] = useState({ firstName:'', lastName:'', email:'', address:'', city:'', zip:'', country:'', cardNumber:'', expiry:'', cvv:'' })

  const set = (k: string) => (e: React.ChangeEvent<HTMLInputElement>) => setForm(f => ({ ...f, [k]: e.target.value }))

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      await api.post('/orders', { items, shippingAddress: form })
      clear()
      router.push('/account/orders')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Checkout failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="container" style={{maxWidth:600}}>
      <h1 className="page-title">Checkout</h1>
      {error && <div className="alert-error">{error}</div>}
      <form onSubmit={submit}>
        <h2 style={{marginBottom:'1rem',fontWeight:600}}>Shipping</h2>
        <div style={{display:'grid',gridTemplateColumns:'1fr 1fr',gap:'1rem'}}>
          <div className="form-group"><label>First Name</label><input required value={form.firstName} onChange={set('firstName')} /></div>
          <div className="form-group"><label>Last Name</label><input required value={form.lastName} onChange={set('lastName')} /></div>
        </div>
        <div className="form-group"><label>Email</label><input type="email" required value={form.email} onChange={set('email')} /></div>
        <div className="form-group"><label>Address</label><input required value={form.address} onChange={set('address')} /></div>
        <div style={{display:'grid',gridTemplateColumns:'2fr 1fr',gap:'1rem'}}>
          <div className="form-group"><label>City</label><input required value={form.city} onChange={set('city')} /></div>
          <div className="form-group"><label>ZIP</label><input required value={form.zip} onChange={set('zip')} /></div>
        </div>
        <div className="form-group"><label>Country</label><input required value={form.country} onChange={set('country')} /></div>

        <h2 style={{margin:'1.5rem 0 1rem',fontWeight:600}}>Payment</h2>
        <div className="form-group"><label>Card Number</label><input required placeholder="4242 4242 4242 4242" value={form.cardNumber} onChange={set('cardNumber')} /></div>
        <div style={{display:'grid',gridTemplateColumns:'1fr 1fr',gap:'1rem'}}>
          <div className="form-group"><label>Expiry</label><input required placeholder="MM/YY" value={form.expiry} onChange={set('expiry')} /></div>
          <div className="form-group"><label>CVV</label><input required placeholder="123" value={form.cvv} onChange={set('cvv')} /></div>
        </div>

        <div style={{border:'1px solid #e5e7eb',borderRadius:'0.5rem',padding:'1rem',margin:'1.5rem 0'}}>
          <div style={{display:'flex',justifyContent:'space-between'}}>
            <span>Order total ({items.length} items)</span>
            <strong>${total().toFixed(2)}</strong>
          </div>
        </div>
        <button type="submit" className="btn-primary" disabled={loading}>
          {loading ? 'Placing order...' : 'Place Order'}
        </button>
      </form>
    </div>
  )
}
