import Link from 'next/link'
import type { Product } from '@/lib/types'

export default function ProductCard({ product }: { product: Product }) {
  return (
    <Link href={`/products/${product.id}`} className="product-card">
      <div className="product-image-wrap">
        <img src={product.imageUrl} alt={product.name} />
      </div>
      <div className="product-info">
        <p className="product-category">{product.category}</p>
        <h3 className="product-name">{product.name}</h3>
        <div className="product-meta">
          <span className="product-price">
            {product.currency} {product.price.toFixed(2)}
          </span>
          <span className="product-rating">{'★'.repeat(Math.round(product.rating))} ({product.reviewCount})</span>
        </div>
        {!product.inStock && <span className="out-of-stock">Out of stock</span>}
      </div>
    </Link>
  )
}
