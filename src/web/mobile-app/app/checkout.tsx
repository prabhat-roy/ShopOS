import { useState } from 'react'
import { View, Text, TextInput, TouchableOpacity, ScrollView, StyleSheet, ActivityIndicator, Alert } from 'react-native'
import { useRouter } from 'expo-router'
import { useCart } from '../src/store/cart'
import { api } from '../src/api/client'

export default function CheckoutScreen() {
  const { items, total, clear } = useCart()
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [form, setForm] = useState({ firstName:'', lastName:'', email:'', address:'', city:'', zip:'', country:'' })
  const set = (k: keyof typeof form) => (v: string) => setForm(f => ({ ...f, [k]: v }))

  async function submit() {
    setLoading(true)
    try {
      await api.post('/orders', { items, shippingAddress: form })
      clear()
      Alert.alert('Order placed!', 'Your order has been placed successfully.', [{ text: 'OK', onPress: () => router.replace('/(tabs)/orders') }])
    } catch { Alert.alert('Error', 'Failed to place order. Please try again.') }
    finally { setLoading(false) }
  }

  return (
    <ScrollView style={s.container} contentContainerStyle={{ padding: 16 }}>
      <Text style={s.title}>Checkout</Text>
      {(['firstName','lastName','email','address','city','zip','country'] as const).map(k => (
        <TextInput key={k} style={s.input} placeholder={k.replace(/([A-Z])/g,' $1').replace(/^./,s=>s.toUpperCase())}
          value={form[k]} onChangeText={set(k)} keyboardType={k==='email'?'email-address':'default'} autoCapitalize={k==='email'?'none':'words'} />
      ))}
      <View style={s.summary}>
        <Text style={s.total}>Total: ${total().toFixed(2)}</Text>
        <Text style={s.items}>{items.length} item(s)</Text>
      </View>
      <TouchableOpacity style={s.btn} onPress={submit} disabled={loading}>
        {loading ? <ActivityIndicator color="#fff" /> : <Text style={s.btnText}>Place Order</Text>}
      </TouchableOpacity>
    </ScrollView>
  )
}

const s = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#fff' },
  title: { fontSize: 22, fontWeight: '700', marginBottom: 16 },
  input: { borderWidth: 1, borderColor: '#d1d5db', borderRadius: 8, padding: 12, marginBottom: 10, fontSize: 15 },
  summary: { borderWidth: 1, borderColor: '#e5e7eb', borderRadius: 8, padding: 14, marginBottom: 16 },
  total: { fontSize: 18, fontWeight: '700' },
  items: { color: '#6b7280', fontSize: 13, marginTop: 2 },
  btn: { backgroundColor: '#111', borderRadius: 8, padding: 16, alignItems: 'center' },
  btnText: { color: '#fff', fontWeight: '700', fontSize: 16 },
})
