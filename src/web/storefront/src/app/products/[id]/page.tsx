import { notFound } from 'next/navigation'
import AddToCart from '@/components/AddToCart'
import { api } from '@/lib/api'
import type { Product } from '@/lib/types'

async function getProduct(id: string): Promise<Product | null> {
  try { return await api.get<Product>(`/products/${id}`) }
  catch { return null }
}

export default async function ProductPage({ params }: { params: { id: string } }) {
  const product = await getProduct(params.id)
  if (!product) notFound()

  return (
    <div className="container">
      <div className="breadcrumb">
        <a href="/">Home</a><span>/</span>
        <a href="/products">Products</a><span>/</span>
        <span>{product.name}</span>
      </div>
      <div className="product-detail">
        <div className="product-detail-image">
          <img src={product.imageUrl} alt={product.name} />
        </div>
        <div className="product-detail-info">
          <p style={{color:'#6b7280',fontSize:'0.875rem'}}>{product.category}</p>
          <h1 style={{fontSize:'1.75rem',fontWeight:700}}>{product.name}</h1>
          <div className="product-detail-rating">
            {'★'.repeat(Math.round(product.rating))} <span style={{color:'#6b7280'}}>({product.reviewCount} reviews)</span>
          </div>
          <p className="product-detail-price">{product.currency} {product.price.toFixed(2)}</p>
          {product.description && <p style={{color:'#374151',lineHeight:1.6}}>{product.description}</p>}
          <AddToCart product={product} />
        </div>
      </div>
    </div>
  )
}
