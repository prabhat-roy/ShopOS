import { useState } from 'react'
import { View, Text, TextInput, TouchableOpacity, StyleSheet, ActivityIndicator } from 'react-native'
import { useRouter } from 'expo-router'
import { useAuth } from '../../src/store/auth'
import { api } from '../../src/api/client'

export default function RegisterScreen() {
  const { setAuth } = useAuth()
  const router = useRouter()
  const [form, setForm] = useState({ firstName: '', lastName: '', email: '', password: '' })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function register() {
    setLoading(true); setError('')
    try {
      const { user, token } = await api.post<{ user: Parameters<typeof setAuth>[1]; token: string }>('/auth/register', form)
      await setAuth(token, user)
      router.replace('/(tabs)')
    } catch { setError('Registration failed') } finally { setLoading(false) }
  }

  return (
    <View style={s.container}>
      <Text style={s.title}>Create Account</Text>
      {error ? <Text style={s.error}>{error}</Text> : null}
      {(['firstName','lastName','email','password'] as const).map(k => (
        <TextInput key={k} style={s.input} placeholder={k.replace(/([A-Z])/g, ' $1').replace(/^./, s => s.toUpperCase())}
          value={form[k]} onChangeText={v => setForm(f => ({...f,[k]:v}))}
          secureTextEntry={k === 'password'} autoCapitalize={k === 'email' ? 'none' : 'words'} keyboardType={k === 'email' ? 'email-address' : 'default'} />
      ))}
      <TouchableOpacity style={s.btn} onPress={register} disabled={loading}>
        {loading ? <ActivityIndicator color="#fff" /> : <Text style={s.btnText}>Register</Text>}
      </TouchableOpacity>
    </View>
  )
}

const s = StyleSheet.create({
  container: { flex: 1, padding: 24, justifyContent: 'center', backgroundColor: '#fff' },
  title: { fontSize: 28, fontWeight: '800', marginBottom: 24 },
  input: { borderWidth: 1, borderColor: '#d1d5db', borderRadius: 8, padding: 12, marginBottom: 12, fontSize: 16 },
  btn: { backgroundColor: '#111', borderRadius: 8, padding: 16, alignItems: 'center', marginBottom: 16 },
  btnText: { color: '#fff', fontWeight: '700', fontSize: 16 },
  error: { color: '#dc2626', marginBottom: 12, fontSize: 14 },
})
