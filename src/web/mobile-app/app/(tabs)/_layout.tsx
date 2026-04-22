import { Tabs } from 'expo-router'

export default function TabLayout() {
  return (
    <Tabs screenOptions={{ tabBarActiveTintColor: '#111' }}>
      <Tabs.Screen name="index"   options={{ title: 'Home'    }} />
      <Tabs.Screen name="search"  options={{ title: 'Search'  }} />
      <Tabs.Screen name="cart"    options={{ title: 'Cart'    }} />
      <Tabs.Screen name="orders"  options={{ title: 'Orders'  }} />
      <Tabs.Screen name="profile" options={{ title: 'Profile' }} />
    </Tabs>
  )
}
