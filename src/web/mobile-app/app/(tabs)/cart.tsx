import { View, Text, FlatList, TouchableOpacity, StyleSheet } from 'react-native'
import { useRouter } from 'expo-router'
import { useCart } from '../../src/store/cart'

export default function CartScreen() {
  const { items, remove, update, total, count } = useCart()
  const router = useRouter()

  if (items.length === 0) return (
    <View style={s.center}>
      <Text style={s.emptyTitle}>Your cart is empty</Text>
      <TouchableOpacity style={s.btn} onPress={() => router.push('/')}><Text style={s.btnText}>Shop Now</Text></TouchableOpacity>
    </View>
  )

  return (
    <View style={s.container}>
      <Text style={s.title}>Cart ({count()} items)</Text>
      <FlatList
        data={items}
        keyExtractor={i => i.product.id}
        renderItem={({ item }) => (
          <View style={s.item}>
            <View style={s.itemInfo}>
              <Text style={s.itemName}>{item.product.name}</Text>
              <Text style={s.itemPrice}>${item.product.price.toFixed(2)}</Text>
            </View>
            <View style={s.qtyRow}>
              <TouchableOpacity style={s.qtyBtn} onPress={() => update(item.product.id, item.quantity - 1)}><Text>−</Text></TouchableOpacity>
              <Text style={s.qty}>{item.quantity}</Text>
              <TouchableOpacity style={s.qtyBtn} onPress={() => update(item.product.id, item.quantity + 1)}><Text>+</Text></TouchableOpacity>
              <TouchableOpacity onPress={() => remove(item.product.id)}><Text style={s.remove}>Remove</Text></TouchableOpacity>
            </View>
          </View>
        )}
      />
      <View style={s.summary}>
        <Text style={s.total}>Total: ${total().toFixed(2)}</Text>
        <TouchableOpacity style={s.btn} onPress={() => router.push('/checkout')}><Text style={s.btnText}>Checkout</Text></TouchableOpacity>
      </View>
    </View>
  )
}

const s = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#fff', padding: 16 },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', gap: 16 },
  title: { fontSize: 22, fontWeight: '700', marginBottom: 16 },
  item: { paddingVertical: 12, borderBottomWidth: 1, borderBottomColor: '#f3f4f6' },
  itemInfo: { flexDirection: 'row', justifyContent: 'space-between', marginBottom: 8 },
  itemName: { fontSize: 15, fontWeight: '500', flex: 1 },
  itemPrice: { fontSize: 15, fontWeight: '700' },
  qtyRow: { flexDirection: 'row', alignItems: 'center', gap: 12 },
  qtyBtn: { width: 28, height: 28, borderWidth: 1, borderColor: '#d1d5db', borderRadius: 4, alignItems: 'center', justifyContent: 'center' },
  qty: { fontSize: 16, fontWeight: '600', minWidth: 20, textAlign: 'center' },
  remove: { color: '#ef4444', fontSize: 13, marginLeft: 8 },
  summary: { paddingTop: 16, gap: 12 },
  total: { fontSize: 18, fontWeight: '700' },
  btn: { backgroundColor: '#111', borderRadius: 8, padding: 14, alignItems: 'center' },
  btnText: { color: '#fff', fontWeight: '600', fontSize: 16 },
  emptyTitle: { fontSize: 18, fontWeight: '600', color: '#374151' },
})
