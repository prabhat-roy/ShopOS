import { useEffect, useState } from 'react'
import { View, Text, FlatList, Image, TouchableOpacity, StyleSheet, ActivityIndicator } from 'react-native'
import { useRouter } from 'expo-router'
import { api } from '../../src/api/client'

interface Product { id: string; name: string; price: number; imageUrl: string; category: string }

export default function HomeScreen() {
  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)
  const router = useRouter()

  useEffect(() => {
    api.get<Product[]>('/products?featured=true&limit=20').then(setProducts).catch(() => {}).finally(() => setLoading(false))
  }, [])

  if (loading) return <View style={s.center}><ActivityIndicator size="large" /></View>

  return (
    <View style={s.container}>
      <Text style={s.title}>Featured Products</Text>
      <FlatList
        data={products}
        numColumns={2}
        keyExtractor={i => i.id}
        columnWrapperStyle={s.row}
        renderItem={({ item }) => (
          <TouchableOpacity style={s.card} onPress={() => router.push(`/product/${item.id}`)}>
            <Image source={{ uri: item.imageUrl }} style={s.img} />
            <Text style={s.name} numberOfLines={2}>{item.name}</Text>
            <Text style={s.price}>${item.price.toFixed(2)}</Text>
          </TouchableOpacity>
        )}
      />
    </View>
  )
}

const s = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#fff', padding: 12 },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center' },
  title: { fontSize: 22, fontWeight: '700', marginBottom: 16 },
  row: { justifyContent: 'space-between' },
  card: { width: '48%', marginBottom: 16, borderRadius: 8, borderWidth: 1, borderColor: '#e5e7eb', overflow: 'hidden' },
  img: { width: '100%', aspectRatio: 1, backgroundColor: '#f3f4f6' },
  name: { padding: 8, fontSize: 13, fontWeight: '500' },
  price: { paddingHorizontal: 8, paddingBottom: 8, fontSize: 14, fontWeight: '700' },
})
