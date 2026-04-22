import { Outlet } from 'react-router-dom'
import Sidebar from './Sidebar'
import Header from './Header'

export default function Layout() {
  return (
    <div style={{display:'flex',height:'100vh',overflow:'hidden'}}>
      <Sidebar />
      <div style={{flex:1,display:'flex',flexDirection:'column',overflow:'hidden'}}>
        <Header />
        <main style={{flex:1,overflow:'auto',padding:'1.5rem',background:'#f9fafb'}}>
          <Outlet />
        </main>
      </div>
    </div>
  )
}
