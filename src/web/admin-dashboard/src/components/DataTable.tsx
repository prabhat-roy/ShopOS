interface Column<T> { key: keyof T; header: string; render?: (v: T[keyof T], row: T) => React.ReactNode }
interface Props<T> { columns: Column<T>[]; data: T[]; keyField: keyof T }

export default function DataTable<T>({ columns, data, keyField }: Props<T>) {
  return (
    <div style={{overflowX:'auto'}}>
      <table style={{width:'100%',borderCollapse:'collapse',fontSize:'0.875rem'}}>
        <thead>
          <tr style={{borderBottom:'2px solid #e5e7eb'}}>
            {columns.map(c => <th key={String(c.key)} style={{textAlign:'left',padding:'0.75rem 1rem',fontWeight:600,color:'#374151'}}>{c.header}</th>)}
          </tr>
        </thead>
        <tbody>
          {data.map(row => (
            <tr key={String(row[keyField])} style={{borderBottom:'1px solid #f3f4f6'}}>
              {columns.map(c => (
                <td key={String(c.key)} style={{padding:'0.75rem 1rem',color:'#374151'}}>
                  {c.render ? c.render(row[c.key], row) : String(row[c.key] ?? '')}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
      {data.length === 0 && <p style={{textAlign:'center',padding:'2rem',color:'#9ca3af'}}>No data</p>}
    </div>
  )
}
