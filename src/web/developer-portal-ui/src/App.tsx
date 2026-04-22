import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import Layout from './components/Layout'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import ApiKeys from './pages/ApiKeys'
import Webhooks from './pages/Webhooks'
import Sandbox from './pages/Sandbox'
import Analytics from './pages/Analytics'
import Docs from './pages/Docs'

const router = createBrowserRouter([
  { path: '/login', element: <Login /> },
  {
    path: '/',
    element: <Layout />,
    children: [
      { index: true,        element: <Dashboard /> },
      { path: 'api-keys',   element: <ApiKeys />   },
      { path: 'webhooks',   element: <Webhooks />  },
      { path: 'sandbox',    element: <Sandbox />   },
      { path: 'analytics',  element: <Analytics /> },
      { path: 'docs',       element: <Docs />      },
    ],
  },
])

export default function App() { return <RouterProvider router={router} /> }
