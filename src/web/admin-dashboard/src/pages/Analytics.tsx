export default function Analytics() {
  return (
    <div>
      <h1 style={{fontSize:'1.5rem',fontWeight:700,marginBottom:'1.5rem'}}>Analytics</h1>
      <div style={{display:'grid',gridTemplateColumns:'1fr 1fr',gap:'1rem'}}>
        {['Revenue Over Time','Orders by Status','Top Products','User Growth'].map(title => (
          <div key={title} style={{background:'#fff',border:'1px solid #e5e7eb',borderRadius:'0.5rem',padding:'1.5rem',minHeight:240}}>
            <h2 style={{fontWeight:600,marginBottom:'1rem',fontSize:'0.875rem',color:'#374151'}}>{title}</h2>
            <div style={{height:160,background:'#f9fafb',borderRadius:'0.25rem',display:'flex',alignItems:'center',justifyContent:'center',color:'#9ca3af',fontSize:'0.875rem'}}>
              Chart — connect to analytics-service
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
