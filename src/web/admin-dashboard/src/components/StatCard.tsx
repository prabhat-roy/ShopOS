interface Props { label: string; value: string | number; change?: number; prefix?: string }
export default function StatCard({ label, value, change, prefix = '' }: Props) {
  return (
    <div style={{background:'#fff',border:'1px solid #e5e7eb',borderRadius:'0.5rem',padding:'1.5rem'}}>
      <p style={{fontSize:'0.75rem',color:'#6b7280',textTransform:'uppercase',letterSpacing:'0.05em',marginBottom:'0.5rem'}}>{label}</p>
      <p style={{fontSize:'1.875rem',fontWeight:700}}>{prefix}{typeof value === 'number' && prefix === '$' ? value.toLocaleString('en-US',{minimumFractionDigits:2}) : value}</p>
      {change !== undefined && (
        <p style={{fontSize:'0.75rem',marginTop:'0.25rem',color: change >= 0 ? '#059669' : '#dc2626'}}>
          {change >= 0 ? '+' : ''}{change.toFixed(1)}% vs last month
        </p>
      )}
    </div>
  )
}
