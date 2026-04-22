import { Outlet, NavLink } from 'react-router-dom'
import { useAuth } from '../store/auth'

const links = [
  { to: '/',          label: 'Dashboard'  },
  { to: '/api-keys',  label: 'API Keys'   },
  { to: '/webhooks',  label: 'Webhooks'   },
  { to: '/sandbox',   label: 'Sandbox'    },
  { to: '/analytics', label: 'Analytics'  },
  { to: '/docs',      label: 'Docs'       },
]

export default function Layout() {
  const { email, logout } = useAuth()
  return (
    <div style={{display:'flex',height:'100vh',fontFamily:'system-ui,sans-serif'}}>
      <aside style={{width:220,background:'#111',color:'#fff',display:'flex',flexDirection:'column',flexShrink:0}}>
        <div style={{padding:'1.5rem',fontWeight:700,fontSize:'1.1rem',borderBottom:'1px solid #374151'}}>Developer Portal</div>
        <nav style={{padding:'0.5rem 0',flex:1}}>
          {links.map(l => (
            <NavLink key={l.to} to={l.to} end={l.to === '/'}
              style={({ isActive }) => ({ display:'block',padding:'0.625rem 1.5rem',color:isActive?'#fff':'#9ca3af',background:isActive?'#374151':'transparent',textDecoration:'none',fontSize:'0.875rem' })}>
              {l.label}
            </NavLink>
          ))}
        </nav>
        <div style={{padding:'1rem 1.5rem',borderTop:'1px solid #374151'}}>
          <p style={{fontSize:'0.75rem',color:'#9ca3af',marginBottom:'0.5rem',overflow:'hidden',textOverflow:'ellipsis',whiteSpace:'nowrap'}}>{email}</p>
          <button onClick={logout} style={{fontSize:'0.75rem',color:'#9ca3af',background:'none',border:'none',cursor:'pointer',padding:0}}>Logout</button>
        </div>
      </aside>
      <div style={{flex:1,display:'flex',flexDirection:'column',overflow:'hidden'}}>
        <header style={{height:60,background:'#fff',borderBottom:'1px solid #e5e7eb',display:'flex',alignItems:'center',padding:'0 1.5rem',flexShrink:0}}>
          <span style={{fontWeight:600,fontSize:'0.875rem',color:'#374151'}}>ShopOS Developer Platform</span>
        </header>
        <main style={{flex:1,overflow:'auto',padding:'1.5rem',background:'#f9fafb'}}><Outlet /></main>
      </div>
    </div>
  )
}
