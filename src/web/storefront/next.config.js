/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  env: {
    API_URL: process.env.API_URL || 'http://web-bff:8081',
  },
  headers: async () => [
    {
      source: '/(.*)',
      headers: [
        { key: 'X-Frame-Options',        value: 'SAMEORIGIN' },
        { key: 'X-Content-Type-Options',  value: 'nosniff' },
        { key: 'Referrer-Policy',         value: 'strict-origin-when-cross-origin' },
        { key: 'Permissions-Policy',      value: 'camera=(), microphone=(), geolocation=()' },
      ],
    },
  ],
  images: {
    domains: ['cdn.shopos.io', 'placehold.co'],
    formats: ['image/avif', 'image/webp'],
  },
}
module.exports = nextConfig
