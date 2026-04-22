'use client'
import Link from 'next/link'
import { useCart } from '@/lib/store/cart'
import { useAuth } from '@/lib/store/auth'

export default function Header() {
  const count = useCart(s => s.count())
  const { user, logout } = useAuth()

  return (
    <header className="header">
      <div className="header-inner">
        <Link href="/" className="logo">ShopOS</Link>
        <nav className="nav">
          <Link href="/products">Products</Link>
          <Link href="/search">Search</Link>
        </nav>
        <div className="header-actions">
          <Link href="/cart" className="cart-btn">
            Cart {count > 0 && <span className="badge">{count}</span>}
          </Link>
          {user ? (
            <>
              <Link href="/account">Account</Link>
              <button onClick={logout}>Logout</button>
            </>
          ) : (
            <Link href="/login">Login</Link>
          )}
        </div>
      </div>
    </header>
  )
}
