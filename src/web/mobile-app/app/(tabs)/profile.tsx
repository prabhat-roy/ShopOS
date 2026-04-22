import { View, Text, TouchableOpacity, StyleSheet } from 'react-native'
import { useRouter } from 'expo-router'
import { useAuth } from '../../src/store/auth'

export default function ProfileScreen() {
  const { user, logout } = useAuth()
  const router = useRouter()

  if (!user) return (
    <View style={s.center}>
      <Text style={s.title}>Sign in to your account</Text>
      <TouchableOpacity style={s.btn} onPress={() => router.push('/(auth)/login')}><Text style={s.btnText}>Login</Text></TouchableOpacity>
      <TouchableOpacity style={s.btnOutline} onPress={() => router.push('/(auth)/register')}><Text style={s.btnOutlineText}>Register</Text></TouchableOpacity>
    </View>
  )

  return (
    <View style={s.container}>
      <Text style={s.title}>My Profile</Text>
      <View style={s.card}>
        <Text style={s.name}>{user.firstName}</Text>
        <Text style={s.email}>{user.email}</Text>
      </View>
      <TouchableOpacity style={s.menuItem} onPress={() => router.push('/(tabs)/orders')}><Text style={s.menuText}>My Orders</Text></TouchableOpacity>
      <TouchableOpacity style={[s.btn, { marginTop: 24 }]} onPress={logout}><Text style={s.btnText}>Logout</Text></TouchableOpacity>
    </View>
  )
}

const s = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#fff', padding: 16 },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', gap: 12, padding: 24 },
  title: { fontSize: 22, fontWeight: '700', marginBottom: 16 },
  card: { borderWidth: 1, borderColor: '#e5e7eb', borderRadius: 8, padding: 16, marginBottom: 16 },
  name: { fontSize: 18, fontWeight: '700' },
  email: { color: '#6b7280', marginTop: 4 },
  menuItem: { padding: 16, borderBottomWidth: 1, borderBottomColor: '#f3f4f6' },
  menuText: { fontSize: 15 },
  btn: { backgroundColor: '#111', borderRadius: 8, padding: 14, alignItems: 'center', width: '100%' },
  btnText: { color: '#fff', fontWeight: '600', fontSize: 16 },
  btnOutline: { borderWidth: 1, borderColor: '#111', borderRadius: 8, padding: 14, alignItems: 'center', width: '100%' },
  btnOutlineText: { fontWeight: '600', fontSize: 16 },
})
