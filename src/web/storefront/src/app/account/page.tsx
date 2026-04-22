'use client'
import Link from 'next/link'
import { useAuth } from '@/lib/store/auth'

export default function AccountPage() {
  const { user } = useAuth()
  if (!user) return null

  return (
    <div className="container" style={{maxWidth:600}}>
      <h1 className="page-title">My Account</h1>
      <div style={{border:'1px solid #e5e7eb',borderRadius:'0.5rem',padding:'1.5rem',marginBottom:'1rem'}}>
        <h2 style={{fontWeight:600,marginBottom:'0.5rem'}}>Profile</h2>
        <p>{user.firstName} {user.lastName}</p>
        <p style={{color:'#6b7280'}}>{user.email}</p>
      </div>
      <div style={{display:'grid',gridTemplateColumns:'1fr 1fr',gap:'1rem'}}>
        <Link href="/account/orders" style={{border:'1px solid #e5e7eb',borderRadius:'0.5rem',padding:'1.5rem',display:'block'}}>
          <h3 style={{fontWeight:600}}>Orders</h3>
          <p style={{color:'#6b7280',fontSize:'0.875rem',marginTop:'0.25rem'}}>View order history</p>
        </Link>
        <Link href="/account/wishlist" style={{border:'1px solid #e5e7eb',borderRadius:'0.5rem',padding:'1.5rem',display:'block'}}>
          <h3 style={{fontWeight:600}}>Wishlist</h3>
          <p style={{color:'#6b7280',fontSize:'0.875rem',marginTop:'0.25rem'}}>Saved items</p>
        </Link>
      </div>
    </div>
  )
}
