import { useState } from 'react'
import { View, Text, TextInput, TouchableOpacity, FlatList, StyleSheet, ActivityIndicator } from 'react-native'
import { useRouter } from 'expo-router'
import { api } from '../../src/api/client'

interface Product { id: string; name: string; price: number; category: string }

export default function SearchScreen() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<Product[]>([])
  const [loading, setLoading] = useState(false)
  const [searched, setSearched] = useState(false)
  const router = useRouter()

  async function search() {
    if (!query.trim()) return
    setLoading(true)
    try { setResults(await api.get<Product[]>(`/search?q=${encodeURIComponent(query)}`)) }
    catch { setResults([]) }
    finally { setLoading(false); setSearched(true) }
  }

  return (
    <View style={s.container}>
      <View style={s.searchRow}>
        <TextInput style={s.input} value={query} onChangeText={setQuery} placeholder="Search products..." onSubmitEditing={search} returnKeyType="search" />
        <TouchableOpacity style={s.btn} onPress={search}><Text style={s.btnText}>Go</Text></TouchableOpacity>
      </View>
      {loading && <ActivityIndicator style={{ marginTop: 20 }} />}
      {searched && !loading && (
        results.length === 0
          ? <Text style={s.empty}>No results for "{query}"</Text>
          : <FlatList data={results} keyExtractor={i => i.id} renderItem={({ item }) => (
              <TouchableOpacity style={s.item} onPress={() => router.push(`/product/${item.id}`)}>
                <Text style={s.itemName}>{item.name}</Text>
                <Text style={s.itemPrice}>${item.price.toFixed(2)}</Text>
              </TouchableOpacity>
            )} />
      )}
    </View>
  )
}

const s = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#fff', padding: 16 },
  searchRow: { flexDirection: 'row', gap: 8, marginBottom: 16 },
  input: { flex: 1, borderWidth: 1, borderColor: '#d1d5db', borderRadius: 8, padding: 10, fontSize: 16 },
  btn: { backgroundColor: '#111', borderRadius: 8, paddingHorizontal: 16, justifyContent: 'center' },
  btnText: { color: '#fff', fontWeight: '600' },
  item: { padding: 14, borderBottomWidth: 1, borderBottomColor: '#f3f4f6', flexDirection: 'row', justifyContent: 'space-between' },
  itemName: { fontSize: 15, fontWeight: '500', flex: 1 },
  itemPrice: { fontSize: 15, fontWeight: '700', color: '#111' },
  empty: { textAlign: 'center', marginTop: 40, color: '#9ca3af' },
})
