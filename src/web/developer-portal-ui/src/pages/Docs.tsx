const ENDPOINTS = [
  { method:'GET',    path:'/products',          desc:'List products with optional filters (category, sort, page)' },
  { method:'GET',    path:'/products/{id}',     desc:'Get a single product by ID' },
  { method:'GET',    path:'/search',            desc:'Full-text search: ?q=query&category=electronics' },
  { method:'POST',   path:'/orders',            desc:'Create a new order with items and shipping address' },
  { method:'GET',    path:'/orders/{id}',       desc:'Get order details by ID' },
  { method:'GET',    path:'/orders/me',         desc:'List orders for the authenticated user' },
  { method:'POST',   path:'/auth/login',        desc:'Authenticate and receive a JWT token' },
  { method:'POST',   path:'/auth/register',     desc:'Create a new user account' },
  { method:'POST',   path:'/cart/items',        desc:'Add item to cart' },
  { method:'DELETE', path:'/cart/items/{id}',   desc:'Remove item from cart' },
]

const METHOD_COLORS: Record<string,string> = { GET:'#dbeafe', POST:'#d1fae5', PUT:'#fef3c7', DELETE:'#fee2e2' }
const METHOD_TEXT: Record<string,string>   = { GET:'#1e40af', POST:'#065f46', PUT:'#92400e', DELETE:'#dc2626' }

export default function Docs() {
  return (
    <div>
      <h1 style={{fontSize:'1.5rem',fontWeight:700,marginBottom:'0.5rem'}}>API Reference</h1>
      <p style={{color:'#6b7280',fontSize:'0.875rem',marginBottom:'1.5rem'}}>Base URL: <code style={{background:'#f3f4f6',padding:'0.1rem 0.4rem',borderRadius:'0.25rem'}}>https://api.shopos.io/v1</code></p>
      <div style={{marginBottom:'1.5rem',background:'#fff',border:'1px solid #e5e7eb',borderRadius:'0.5rem',padding:'1rem'}}>
        <p style={{fontWeight:600,marginBottom:'0.5rem',fontSize:'0.875rem'}}>Authentication</p>
        <p style={{color:'#374151',fontSize:'0.875rem'}}>All API calls require an API key passed in the Authorization header:</p>
        <code style={{display:'block',background:'#f9fafb',padding:'0.75rem',borderRadius:'0.375rem',fontFamily:'monospace',fontSize:'0.8rem',marginTop:'0.5rem'}}>Authorization: Bearer sk_live_xxxxxxxxxxxx</code>
      </div>
      <div style={{background:'#fff',border:'1px solid #e5e7eb',borderRadius:'0.5rem',overflow:'hidden'}}>
        {ENDPOINTS.map((ep,i) => (
          <div key={i} style={{padding:'1rem 1.5rem',borderBottom:'1px solid #f3f4f6',display:'flex',alignItems:'flex-start',gap:'1rem'}}>
            <span style={{padding:'0.2rem 0.5rem',borderRadius:'0.25rem',fontSize:'0.7rem',fontWeight:700,background:METHOD_COLORS[ep.method],color:METHOD_TEXT[ep.method],flexShrink:0,minWidth:52,textAlign:'center'}}>{ep.method}</span>
            <div>
              <code style={{fontFamily:'monospace',fontSize:'0.875rem',fontWeight:600}}>{ep.path}</code>
              <p style={{color:'#6b7280',fontSize:'0.8rem',marginTop:'0.25rem'}}>{ep.desc}</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
