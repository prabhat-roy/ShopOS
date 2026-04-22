import Link from 'next/link'
import ProductCard from '@/components/ProductCard'
import { api } from '@/lib/api'
import type { Product } from '@/lib/types'

async function getFeatured(): Promise<Product[]> {
  try { return await api.get<Product[]>('/products?featured=true&limit=8') }
  catch { return [] }
}

export default async function HomePage() {
  const products = await getFeatured()
  return (
    <>
      <section className="hero">
        <div className="container">
          <h1>Welcome to ShopOS</h1>
          <p>Discover thousands of products across every category.</p>
          <div className="hero-actions">
            <Link href="/products" className="btn-primary" style={{display:'inline-block',padding:'0.75rem 2rem'}}>Shop Now</Link>
            <Link href="/search" className="btn-secondary" style={{display:'inline-block',padding:'0.75rem 2rem'}}>Search</Link>
          </div>
        </div>
      </section>

      <section className="section">
        <div className="container">
          <h2 className="section-title">Featured Products</h2>
          {products.length > 0 ? (
            <div className="product-grid">
              {products.map(p => <ProductCard key={p.id} product={p} />)}
            </div>
          ) : (
            <p style={{color:'#6b7280'}}>No featured products available.</p>
          )}
        </div>
      </section>
    </>
  )
}
