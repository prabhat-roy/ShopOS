import { useState } from 'react'
export default function Settings() {
  const [saved, setSaved] = useState(false)
  return (
    <div style={{maxWidth:600}}>
      <h1 style={{fontSize:'1.5rem',fontWeight:700,marginBottom:'1.5rem'}}>Settings</h1>
      {saved && <div style={{background:'#d1fae5',color:'#065f46',padding:'0.75rem 1rem',borderRadius:'0.375rem',marginBottom:'1rem'}}>Settings saved.</div>}
      <div style={{background:'#fff',border:'1px solid #e5e7eb',borderRadius:'0.5rem',padding:'1.5rem',marginBottom:'1rem'}}>
        <h2 style={{fontWeight:600,marginBottom:'1rem'}}>General</h2>
        {['Platform Name','Support Email','Default Currency'].map(label => (
          <div key={label} style={{marginBottom:'1rem'}}>
            <label style={{display:'block',fontSize:'0.875rem',fontWeight:500,marginBottom:'0.25rem'}}>{label}</label>
            <input style={{width:'100%',border:'1px solid #d1d5db',borderRadius:'0.375rem',padding:'0.5rem 0.75rem'}} />
          </div>
        ))}
        <button onClick={() => setSaved(true)} style={{background:'#111',color:'#fff',border:'none',borderRadius:'0.375rem',padding:'0.5rem 1.5rem',cursor:'pointer'}}>Save Changes</button>
      </div>
    </div>
  )
}
