'use client'
import Link from 'next/link'
import { useCart } from '@/lib/store/cart'

export default function CartPage() {
  const { items, remove, update, total } = useCart()

  if (items.length === 0) {
    return (
      <div className="container">
        <div className="empty-state">
          <h2>Your cart is empty</h2>
          <p>Start shopping to add items.</p>
          <Link href="/products" className="btn-primary" style={{display:'inline-block',marginTop:'1rem',padding:'0.75rem 1.5rem'}}>Browse Products</Link>
        </div>
      </div>
    )
  }

  return (
    <div className="container" style={{maxWidth:800}}>
      <h1 className="page-title">Shopping Cart</h1>
      {items.map(item => (
        <div key={item.product.id} className="cart-item">
          <img className="cart-item-image" src={item.product.imageUrl} alt={item.product.name} />
          <div className="cart-item-details">
            <Link href={`/products/${item.product.id}`}><strong>{item.product.name}</strong></Link>
            <p style={{color:'#6b7280',fontSize:'0.875rem'}}>{item.product.currency} {item.product.price.toFixed(2)}</p>
            <div style={{display:'flex',alignItems:'center',gap:'0.5rem',marginTop:'0.5rem'}}>
              <button className="btn-secondary" onClick={() => update(item.product.id, item.quantity - 1)}>−</button>
              <span>{item.quantity}</span>
              <button className="btn-secondary" onClick={() => update(item.product.id, item.quantity + 1)}>+</button>
              <button style={{marginLeft:'auto',color:'#ef4444',background:'none',border:'none',cursor:'pointer'}} onClick={() => remove(item.product.id)}>Remove</button>
            </div>
          </div>
          <div style={{fontWeight:700,whiteSpace:'nowrap'}}>
            {item.product.currency} {(item.product.price * item.quantity).toFixed(2)}
          </div>
        </div>
      ))}
      <div className="cart-summary">
        <div style={{display:'flex',justifyContent:'space-between',marginBottom:'1rem'}}>
          <span style={{fontWeight:600}}>Total</span>
          <span style={{fontWeight:700,fontSize:'1.25rem'}}>${total().toFixed(2)}</span>
        </div>
        <Link href="/checkout" className="btn-primary" style={{display:'block',textAlign:'center'}}>Proceed to Checkout</Link>
      </div>
    </div>
  )
}
