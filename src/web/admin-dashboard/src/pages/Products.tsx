import { useEffect, useState } from 'react'
import DataTable from '../components/DataTable'
import { productsApi } from '../api/client'
import type { Product } from '../types'

export default function Products() {
  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => { productsApi.list().then(setProducts).catch(()=>{}).finally(()=>setLoading(false)) }, [])

  async function del(id: string) {
    if (!confirm('Delete product?')) return
    await productsApi.delete(id)
    setProducts(p => p.filter(x => x.id !== id))
  }

  return (
    <div>
      <div style={{display:'flex',justifyContent:'space-between',alignItems:'center',marginBottom:'1.5rem'}}>
        <h1 style={{fontSize:'1.5rem',fontWeight:700}}>Products</h1>
        <button style={{background:'#111',color:'#fff',border:'none',borderRadius:'0.375rem',padding:'0.5rem 1rem',cursor:'pointer',fontSize:'0.875rem'}}>+ Add Product</button>
      </div>
      <div style={{background:'#fff',border:'1px solid #e5e7eb',borderRadius:'0.5rem',overflow:'hidden'}}>
        {loading ? <p style={{padding:'2rem',textAlign:'center',color:'#9ca3af'}}>Loading...</p> : (
          <DataTable<Product>
            keyField="id"
            data={products}
            columns={[
              { key:'name',     header:'Name' },
              { key:'category', header:'Category' },
              { key:'price',    header:'Price',  render: v => `$${Number(v).toFixed(2)}` },
              { key:'stock',    header:'Stock',  render: v => <span style={{color: Number(v) < 10 ? '#dc2626' : '#059669'}}>{String(v)}</span> },
              { key:'id',       header:'Actions', render: (_,row) => <button onClick={() => del(row.id)} style={{color:'#dc2626',background:'none',border:'none',cursor:'pointer',fontSize:'0.875rem'}}>Delete</button> },
            ]}
          />
        )}
      </div>
    </div>
  )
}
