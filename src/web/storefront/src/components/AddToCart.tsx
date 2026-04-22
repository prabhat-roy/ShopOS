'use client'
import { useCart } from '@/lib/store/cart'
import type { Product } from '@/lib/types'

export default function AddToCart({ product }: { product: Product }) {
  const add = useCart(s => s.add)
  return (
    <button
      className="btn-primary"
      disabled={!product.inStock}
      onClick={() => add(product)}
    >
      {product.inStock ? 'Add to Cart' : 'Out of Stock'}
    </button>
  )
}
