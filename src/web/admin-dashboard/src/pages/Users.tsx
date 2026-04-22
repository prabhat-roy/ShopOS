import { useEffect, useState } from 'react'
import DataTable from '../components/DataTable'
import { usersApi } from '../api/client'
import type { User } from '../types'

export default function Users() {
  const [users, setUsers] = useState<User[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => { usersApi.list().then(setUsers).catch(()=>{}).finally(()=>setLoading(false)) }, [])

  return (
    <div>
      <h1 style={{fontSize:'1.5rem',fontWeight:700,marginBottom:'1.5rem'}}>Users</h1>
      <div style={{background:'#fff',border:'1px solid #e5e7eb',borderRadius:'0.5rem',overflow:'hidden'}}>
        {loading ? <p style={{padding:'2rem',textAlign:'center',color:'#9ca3af'}}>Loading...</p> : (
          <DataTable<User>
            keyField="id"
            data={users}
            columns={[
              { key:'email',     header:'Email' },
              { key:'firstName', header:'First Name' },
              { key:'lastName',  header:'Last Name' },
              { key:'status',    header:'Status',     render: v => <span style={{padding:'0.2rem 0.5rem',borderRadius:'9999px',fontSize:'0.75rem',fontWeight:600,background: v==='active'?'#d1fae5':'#fee2e2',color: v==='active'?'#065f46':'#dc2626'}}>{String(v)}</span> },
              { key:'createdAt', header:'Joined',      render: v => new Date(String(v)).toLocaleDateString() },
            ]}
          />
        )}
      </div>
    </div>
  )
}
