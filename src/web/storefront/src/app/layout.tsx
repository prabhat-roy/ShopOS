import type { Metadata } from 'next'
import Header from '@/components/Header'
import Footer from '@/components/Footer'
import Providers from '@/components/Providers'
import './globals.css'

export const metadata: Metadata = {
  title: { default: 'ShopOS', template: '%s | ShopOS' },
  description: 'Enterprise commerce platform',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <Providers>
          <Header />
          <main className="main-content">{children}</main>
          <Footer />
        </Providers>
      </body>
    </html>
  )
}
