import { NavLink } from 'react-router-dom'

const links = [
  { to: '/',          label: 'Dashboard' },
  { to: '/orders',    label: 'Orders'    },
  { to: '/products',  label: 'Products'  },
  { to: '/users',     label: 'Users'     },
  { to: '/analytics', label: 'Analytics' },
  { to: '/settings',  label: 'Settings'  },
]

export default function Sidebar() {
  return (
    <aside style={{width:220,background:'#111',color:'#fff',display:'flex',flexDirection:'column',flexShrink:0}}>
      <div style={{padding:'1.5rem',fontWeight:700,fontSize:'1.25rem',borderBottom:'1px solid #374151'}}>ShopOS Admin</div>
      <nav style={{padding:'1rem 0',flex:1}}>
        {links.map(l => (
          <NavLink key={l.to} to={l.to} end={l.to === '/'}
            style={({ isActive }) => ({
              display:'block', padding:'0.625rem 1.5rem', color: isActive ? '#fff' : '#9ca3af',
              background: isActive ? '#374151' : 'transparent', textDecoration:'none', fontSize:'0.875rem',
            })}>
            {l.label}
          </NavLink>
        ))}
      </nav>
    </aside>
  )
}
