import { useState } from 'react'
import { View, Text, TextInput, TouchableOpacity, StyleSheet, ActivityIndicator } from 'react-native'
import { useRouter } from 'expo-router'
import { useAuth } from '../../src/store/auth'
import { api } from '../../src/api/client'

export default function LoginScreen() {
  const { setAuth } = useAuth()
  const router = useRouter()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function login() {
    setLoading(true); setError('')
    try {
      const { user, token } = await api.post<{ user: Parameters<typeof setAuth>[1]; token: string }>('/auth/login', { email, password })
      await setAuth(token, user)
      router.replace('/(tabs)')
    } catch { setError('Invalid email or password') } finally { setLoading(false) }
  }

  return (
    <View style={s.container}>
      <Text style={s.title}>Login</Text>
      {error ? <Text style={s.error}>{error}</Text> : null}
      <TextInput style={s.input} placeholder="Email" value={email} onChangeText={setEmail} keyboardType="email-address" autoCapitalize="none" />
      <TextInput style={s.input} placeholder="Password" value={password} onChangeText={setPassword} secureTextEntry />
      <TouchableOpacity style={s.btn} onPress={login} disabled={loading}>
        {loading ? <ActivityIndicator color="#fff" /> : <Text style={s.btnText}>Login</Text>}
      </TouchableOpacity>
      <TouchableOpacity onPress={() => router.push('/(auth)/register')}><Text style={s.link}>Don't have an account? Register</Text></TouchableOpacity>
    </View>
  )
}

const s = StyleSheet.create({
  container: { flex: 1, padding: 24, justifyContent: 'center', backgroundColor: '#fff' },
  title: { fontSize: 28, fontWeight: '800', marginBottom: 24 },
  input: { borderWidth: 1, borderColor: '#d1d5db', borderRadius: 8, padding: 12, marginBottom: 12, fontSize: 16 },
  btn: { backgroundColor: '#111', borderRadius: 8, padding: 16, alignItems: 'center', marginBottom: 16 },
  btnText: { color: '#fff', fontWeight: '700', fontSize: 16 },
  link: { textAlign: 'center', color: '#6b7280' },
  error: { color: '#dc2626', marginBottom: 12, fontSize: 14 },
})
