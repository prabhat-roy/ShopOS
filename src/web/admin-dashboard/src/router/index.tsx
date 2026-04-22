import { createBrowserRouter } from 'react-router-dom'
import Layout from '../components/Layout'
import Login from '../pages/Login'
import Dashboard from '../pages/Dashboard'
import Orders from '../pages/Orders'
import Products from '../pages/Products'
import Users from '../pages/Users'
import Analytics from '../pages/Analytics'
import Settings from '../pages/Settings'

export const router = createBrowserRouter([
  { path: '/login', element: <Login /> },
  {
    path: '/',
    element: <Layout />,
    children: [
      { index: true,        element: <Dashboard /> },
      { path: 'orders',     element: <Orders /> },
      { path: 'products',   element: <Products /> },
      { path: 'users',      element: <Users /> },
      { path: 'analytics',  element: <Analytics /> },
      { path: 'settings',   element: <Settings /> },
    ],
  },
])
