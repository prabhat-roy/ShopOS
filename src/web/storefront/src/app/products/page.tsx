import ProductCard from '@/components/ProductCard'
import { api } from '@/lib/api'
import type { Product } from '@/lib/types'

interface Props { searchParams: { category?: string; sort?: string; page?: string } }

async function getProducts(params: Props['searchParams']): Promise<Product[]> {
  const q = new URLSearchParams()
  if (params.category) q.set('category', params.category)
  if (params.sort) q.set('sort', params.sort)
  q.set('page', params.page ?? '1')
  try { return await api.get<Product[]>(`/products?${q}`) }
  catch { return [] }
}

export default async function ProductsPage({ searchParams }: Props) {
  const products = await getProducts(searchParams)
  return (
    <div className="container">
      <h1 className="page-title">All Products</h1>
      <div className="filter-bar">
        <select defaultValue={searchParams.category ?? ''} name="category">
          <option value="">All Categories</option>
          <option value="electronics">Electronics</option>
          <option value="clothing">Clothing</option>
          <option value="home">Home & Garden</option>
          <option value="books">Books</option>
        </select>
        <select defaultValue={searchParams.sort ?? 'newest'} name="sort">
          <option value="newest">Newest</option>
          <option value="price_asc">Price: Low to High</option>
          <option value="price_desc">Price: High to Low</option>
          <option value="rating">Top Rated</option>
        </select>
      </div>
      {products.length > 0 ? (
        <div className="product-grid">
          {products.map(p => <ProductCard key={p.id} product={p} />)}
        </div>
      ) : (
        <div className="empty-state"><h2>No products found</h2><p>Try adjusting your filters.</p></div>
      )}
    </div>
  )
}
