import Link from 'next/link'

export default function NotFound() {
  return (
    <div className="container" style={{textAlign:'center',paddingTop:'5rem'}}>
      <h1 style={{fontSize:'4rem',fontWeight:800,color:'#e5e7eb'}}>404</h1>
      <h2 style={{fontSize:'1.5rem',fontWeight:700,marginBottom:'0.5rem'}}>Page not found</h2>
      <p style={{color:'#6b7280',marginBottom:'2rem'}}>The page you're looking for doesn't exist.</p>
      <Link href="/" className="btn-primary" style={{display:'inline-block',padding:'0.75rem 1.5rem'}}>Go Home</Link>
    </div>
  )
}
