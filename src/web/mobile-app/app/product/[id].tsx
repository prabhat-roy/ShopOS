import { useEffect, useState } from 'react'
import { View, Text, Image, TouchableOpacity, ScrollView, StyleSheet, ActivityIndicator } from 'react-native'
import { useLocalSearchParams } from 'expo-router'
import { useCart } from '../../src/store/cart'
import { api } from '../../src/api/client'

interface Product { id: string; name: string; price: number; imageUrl: string; category: string; description?: string; inStock: boolean; rating: number }

export default function ProductScreen() {
  const { id } = useLocalSearchParams<{ id: string }>()
  const [product, setProduct] = useState<Product | null>(null)
  const [loading, setLoading] = useState(true)
  const add = useCart(s => s.add)
  const [added, setAdded] = useState(false)

  useEffect(() => { api.get<Product>(`/products/${id}`).then(setProduct).catch(()=>{}).finally(()=>setLoading(false)) }, [id])

  function addToCart() { if (product) { add(product); setAdded(true); setTimeout(() => setAdded(false), 1500) } }

  if (loading) return <View style={s.center}><ActivityIndicator size="large" /></View>
  if (!product) return <View style={s.center}><Text>Product not found.</Text></View>

  return (
    <ScrollView style={s.container}>
      <Image source={{ uri: product.imageUrl }} style={s.img} />
      <View style={s.info}>
        <Text style={s.category}>{product.category}</Text>
        <Text style={s.name}>{product.name}</Text>
        <Text style={s.rating}>{'★'.repeat(Math.round(product.rating))} {product.rating.toFixed(1)}</Text>
        <Text style={s.price}>${product.price.toFixed(2)}</Text>
        {product.description && <Text style={s.desc}>{product.description}</Text>}
        <TouchableOpacity style={[s.btn, !product.inStock && s.btnDisabled]} onPress={addToCart} disabled={!product.inStock}>
          <Text style={s.btnText}>{!product.inStock ? 'Out of Stock' : added ? 'Added!' : 'Add to Cart'}</Text>
        </TouchableOpacity>
      </View>
    </ScrollView>
  )
}

const s = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#fff' },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center' },
  img: { width: '100%', aspectRatio: 1, backgroundColor: '#f3f4f6' },
  info: { padding: 16, gap: 8 },
  category: { color: '#6b7280', fontSize: 12, textTransform: 'uppercase', letterSpacing: 1 },
  name: { fontSize: 22, fontWeight: '700' },
  rating: { color: '#f59e0b', fontSize: 14 },
  price: { fontSize: 24, fontWeight: '800' },
  desc: { color: '#374151', lineHeight: 22 },
  btn: { backgroundColor: '#111', borderRadius: 8, padding: 16, alignItems: 'center', marginTop: 8 },
  btnDisabled: { backgroundColor: '#9ca3af' },
  btnText: { color: '#fff', fontWeight: '700', fontSize: 16 },
})
