/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./app/**/*.{ts,tsx}', './src/**/*.{ts,tsx}'],
  presets: [require('nativewind/preset')],
  theme: {
    extend: {
      colors: {
        primary: '#111111',
        secondary: '#6b7280',
        accent: '#3b82f6',
        success: '#059669',
        danger: '#dc2626',
        warning: '#f59e0b',
      },
    },
  },
  plugins: [],
}
