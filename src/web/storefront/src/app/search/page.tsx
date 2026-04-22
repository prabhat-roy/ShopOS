'use client'
import { useState } from 'react'
import ProductCard from '@/components/ProductCard'
import { api } from '@/lib/api'
import type { Product } from '@/lib/types'

export default function SearchPage() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<Product[]>([])
  const [loading, setLoading] = useState(false)
  const [searched, setSearched] = useState(false)

  async function search(e: React.FormEvent) {
    e.preventDefault()
    if (!query.trim()) return
    setLoading(true)
    try {
      const data = await api.get<Product[]>(`/search?q=${encodeURIComponent(query)}`)
      setResults(data)
    } catch { setResults([]) }
    finally { setLoading(false); setSearched(true) }
  }

  return (
    <div className="container">
      <h1 className="page-title">Search</h1>
      <form onSubmit={search} style={{display:'flex',gap:'0.5rem',marginBottom:'2rem'}}>
        <input
          style={{flex:1,border:'1px solid #d1d5db',borderRadius:'0.375rem',padding:'0.5rem 0.75rem',fontSize:'1rem'}}
          value={query} onChange={e => setQuery(e.target.value)}
          placeholder="Search products..."
        />
        <button type="submit" className="btn-primary" style={{width:'auto',padding:'0.5rem 1.5rem'}} disabled={loading}>
          {loading ? '...' : 'Search'}
        </button>
      </form>
      {searched && (
        results.length > 0 ? (
          <>
            <p style={{color:'#6b7280',marginBottom:'1rem'}}>{results.length} results for "{query}"</p>
            <div className="product-grid">{results.map(p => <ProductCard key={p.id} product={p} />)}</div>
          </>
        ) : (
          <div className="empty-state"><h2>No results found</h2><p>Try a different search term.</p></div>
        )
      )}
    </div>
  )
}
