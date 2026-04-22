import { useAuth } from '../store/auth'

export default function Header() {
  const { email, logout } = useAuth()
  return (
    <header style={{height:60,background:'#fff',borderBottom:'1px solid #e5e7eb',display:'flex',alignItems:'center',padding:'0 1.5rem',justifyContent:'space-between',flexShrink:0}}>
      <div />
      <div style={{display:'flex',alignItems:'center',gap:'1rem'}}>
        <span style={{fontSize:'0.875rem',color:'#6b7280'}}>{email}</span>
        <button onClick={logout} style={{fontSize:'0.875rem',color:'#ef4444',background:'none',border:'none',cursor:'pointer'}}>Logout</button>
      </div>
    </header>
  )
}
