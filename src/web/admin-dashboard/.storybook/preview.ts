import type { Preview } from '@storybook/react'
const preview: Preview = {
  parameters: {
    controls: { matchers: { color: /(background|color)$/i, date: /Date$/i } },
    backgrounds: { default: 'light', values: [{ name: 'light', value: '#ffffff' }, { name: 'dark', value: '#111111' }] },
  },
}
export default preview
