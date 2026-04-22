import { useEffect, useState } from 'react'
import { View, Text, FlatList, StyleSheet, ActivityIndicator } from 'react-native'
import { api } from '../../src/api/client'

interface Order { id: string; status: string; total: number; createdAt: string; items: number }

export default function OrdersScreen() {
  const [orders, setOrders] = useState<Order[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => { api.get<Order[]>('/orders/me').then(setOrders).catch(()=>{}).finally(()=>setLoading(false)) }, [])

  if (loading) return <View style={s.center}><ActivityIndicator size="large" /></View>

  return (
    <View style={s.container}>
      <Text style={s.title}>My Orders</Text>
      {orders.length === 0 ? <Text style={s.empty}>No orders yet.</Text> : (
        <FlatList data={orders} keyExtractor={o => o.id} renderItem={({ item: o }) => (
          <View style={s.card}>
            <View style={s.row}>
              <Text style={s.id}>#{o.id.slice(0,8).toUpperCase()}</Text>
              <Text style={[s.status, { color: o.status === 'delivered' ? '#059669' : '#374151' }]}>{o.status}</Text>
            </View>
            <Text style={s.meta}>{o.items} item(s) · ${o.total.toFixed(2)}</Text>
            <Text style={s.date}>{new Date(o.createdAt).toLocaleDateString()}</Text>
          </View>
        )} />
      )}
    </View>
  )
}

const s = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#fff', padding: 16 },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center' },
  title: { fontSize: 22, fontWeight: '700', marginBottom: 16 },
  card: { borderWidth: 1, borderColor: '#e5e7eb', borderRadius: 8, padding: 14, marginBottom: 12 },
  row: { flexDirection: 'row', justifyContent: 'space-between', marginBottom: 4 },
  id: { fontWeight: '700', fontSize: 15 },
  status: { fontWeight: '600', fontSize: 13 },
  meta: { color: '#6b7280', fontSize: 13, marginBottom: 2 },
  date: { color: '#9ca3af', fontSize: 12 },
  empty: { textAlign: 'center', marginTop: 40, color: '#9ca3af' },
})
